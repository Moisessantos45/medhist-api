package appointments

import (
	"api_citas/internal/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type AppointmentUseCase struct {
	repo models.AppointmentRepository
	rd   *redis.Client
}

const appointmentCacheKeyPrefix = "appointment:"

func NewAppointmentUseCase(repo models.AppointmentRepository, rd *redis.Client) models.AppointmentUseCase {
	return &AppointmentUseCase{
		repo: repo,
		rd:   rd,
	}
}

func (r *AppointmentUseCase) GetAll(ctx context.Context, patientID uint64, veterinarianID uint64, page int, pageSize int) (*models.PaginatedAppointments, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	cacheKey := fmt.Sprintf("%sall:%d:%d:%d:%d", appointmentCacheKeyPrefix, patientID, veterinarianID, page, pageSize)

	cachedData, err := r.rd.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var result models.PaginatedAppointments
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return &result, nil
		}
	} else if err != redis.Nil {
		log.Printf("Redis error: %v", err)
	}

	appointments, total, err := r.repo.GetAll(patientID, veterinarianID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("error fetching appointments: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	if page > totalPages && total > 0 {
		return nil, fmt.Errorf("page %d exceeds total pages %d (total items: %d)", page, totalPages, total)
	}

	result := &models.PaginatedAppointments{
		Data: appointments,
		Paginate: models.Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	go func() {
		data, err := json.Marshal(result)
		if err != nil {
			log.Printf("Marshal error: %v", err)
			return
		}

		if err := r.rd.Set(ctx, cacheKey, data, 5*time.Minute).Err(); err != nil {
			log.Printf("Cache set error: %v", err)
		}
	}()

	return result, nil
}

func (r *AppointmentUseCase) GetByID(id uint64) (*models.Appointment, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid id: %d", id)
	}

	return r.repo.GetByID(id)
}

func (r *AppointmentUseCase) Create(ctx context.Context, a *models.Appointment) error {
	newAppointment, err := models.NewAppointment(a.Date, a.Reason, a.Status, a.Notes, a.PatientID, a.VeterinarianID)
	if err != nil {
		return err
	}

	err = r.repo.Create(newAppointment)
	if err != nil {
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key value") {
			return fmt.Errorf("la cita ya se encuentra registrada")
		}
		return err
	}

	*a = *newAppointment

	go func(patientID uint64, vetID uint64) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pattern := fmt.Sprintf("%sall:%d:%d:*", appointmentCacheKeyPrefix, patientID, vetID)
		iter := r.rd.Scan(bgCtx, 0, pattern, 0).Iterator()

		for iter.Next(bgCtx) {
			if err := r.rd.Del(bgCtx, iter.Val()).Err(); err != nil {
				log.Printf("Error invalidando caché %s: %v", iter.Val(), err)
			}
		}
	}(a.PatientID, a.VeterinarianID)

	return nil
}

func (r *AppointmentUseCase) Update(ctx context.Context, id uint64, a *models.Appointment) error {
	if id == 0 {
		return fmt.Errorf("invalid id: %d", id)
	}

	updatedAppointment, err := models.NewAppointment(a.Date, a.Reason, a.Status, a.Notes, a.PatientID, a.VeterinarianID)
	if err != nil {
		return err
	}

	err = r.repo.Update(id, updatedAppointment)
	if err != nil {
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key value") {
			return fmt.Errorf("la cita ya se encuentra registrada")
		}
		return err
	}

	*a = *updatedAppointment

	go func(patientID uint64, vetID uint64) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pattern := fmt.Sprintf("%sall:%d:%d:*", appointmentCacheKeyPrefix, patientID, vetID)
		iter := r.rd.Scan(bgCtx, 0, pattern, 0).Iterator()

		for iter.Next(bgCtx) {
			if err := r.rd.Del(bgCtx, iter.Val()).Err(); err != nil {
				log.Printf("Error invalidando caché %s: %v", iter.Val(), err)
			}
		}
	}(a.PatientID, a.VeterinarianID)

	return nil
}

func (r *AppointmentUseCase) Delete(id uint64, veterinarianID uint64) error {
	if id == 0 {
		return fmt.Errorf("invalid id: %d", id)
	}

	return r.repo.Delete(id, veterinarianID)
}
