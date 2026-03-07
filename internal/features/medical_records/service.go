package medicalrecords

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

type MedicalRecordUseCase struct {
	repo models.MedicalRecordRepository
	rd   *redis.Client
}

const medicalRecordCacheKeyPrefix = "medical_record:"

func NewMedicalRecordUseCase(repo models.MedicalRecordRepository, rd *redis.Client) models.MedicalRecordUseCase {
	return &MedicalRecordUseCase{
		repo: repo,
		rd:   rd,
	}
}

func (s *MedicalRecordUseCase) GetAll(ctx context.Context, id uint64, veterinarianId uint64, page int, pageSize int) (*models.PaginatedMedicalRecords, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	cacheKey := fmt.Sprintf("%sall:%d:%d:%d:%d", medicalRecordCacheKeyPrefix, id, veterinarianId, page, pageSize)

	cachedData, err := s.rd.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var result *models.PaginatedMedicalRecords
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return result, nil
		}
	} else if err != redis.Nil {
		fmt.Printf("Redis error: %v\n", err)
	}

	results, total, err := s.repo.GetAll(id, veterinarianId, offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if page > totalPages && total > 0 {
		return nil, fmt.Errorf("page %d exceeds total pages %d (total items: %d)", page, totalPages, total)
	}

	result := &models.PaginatedMedicalRecords{
		Data: results,
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
			fmt.Printf("Marshal error: %v\n", err)
			return
		}

		if err := s.rd.Set(ctx, cacheKey, data, 5*time.Minute).Err(); err != nil {
			fmt.Printf("Redis set error: %v\n", err)
		}

	}()

	return result, nil
}

func (s *MedicalRecordUseCase) GetByID(id uint64) (*models.MedicalRecord, error) {
	if id == 0 {
		return nil, fmt.Errorf("id is required")
	}

	return s.repo.GetByID(id)
}

func (s *MedicalRecordUseCase) Create(ctx context.Context, mr *models.MedicalRecord) error {
	newMR, err := models.NewMedicalRecord(mr.VisitDate, mr.Diagnosis, mr.Treatment, mr.Prescription, mr.WeightKg, mr.TemperatureC, mr.Notes, mr.PatientID, mr.VeterinarianID)

	if err != nil {
		return err
	}

	err = s.repo.Create(newMR)
	if err != nil {
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key value") {
			return fmt.Errorf("el registro médico ya se encuentra registrado")
		}
		return err
	}

	*mr = *newMR

	pattern := fmt.Sprintf("%sall:*:%d:%d:*", medicalRecordCacheKeyPrefix, mr.PatientID, mr.VeterinarianID)
	iter := s.rd.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		s.rd.Del(ctx, iter.Val())
	}

	go func(patientId uint64, veterinarianId uint64) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pattern := fmt.Sprintf("%sall:%d:%d:*", medicalRecordCacheKeyPrefix, patientId, veterinarianId)
		iter := s.rd.Scan(bgCtx, 0, pattern, 0).Iterator()

		for iter.Next(bgCtx) {
			if err := s.rd.Del(bgCtx, iter.Val()).Err(); err != nil {
				log.Printf("Error invalidando caché %s: %v", iter.Val(), err)
			}
		}

	}(mr.PatientID, mr.VeterinarianID)

	return nil
}

func (s *MedicalRecordUseCase) Update(ctx context.Context, id uint64, mr *models.MedicalRecord) error {
	if id == 0 {
		return fmt.Errorf("id is required")
	}

	updatedMR, err := models.NewMedicalRecord(mr.VisitDate, mr.Diagnosis, mr.Treatment, mr.Prescription, mr.WeightKg, mr.TemperatureC, mr.Notes, mr.PatientID, mr.VeterinarianID)

	if err != nil {
		return err
	}

	err = s.repo.Update(id, updatedMR)
	if err != nil {
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key value") {
			return fmt.Errorf("el registro médico ya se encuentra registrado")
		}
		return err
	}

	*mr = *updatedMR

	go func(patientId uint64, veterinarianId uint64) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pattern := fmt.Sprintf("%sall:%d:%d:*", medicalRecordCacheKeyPrefix, patientId, veterinarianId)
		iter := s.rd.Scan(bgCtx, 0, pattern, 0).Iterator()

		for iter.Next(bgCtx) {
			if err := s.rd.Del(bgCtx, iter.Val()).Err(); err != nil {
				log.Printf("Error invalidando caché %s: %v", iter.Val(), err)
			}
		}

	}(mr.PatientID, mr.VeterinarianID)

	return nil
}

func (s *MedicalRecordUseCase) Delete(id uint64, veterinarianID uint64) error {
	if id == 0 {
		return fmt.Errorf("id is required")
	}

	return s.repo.Delete(id, veterinarianID)
}
