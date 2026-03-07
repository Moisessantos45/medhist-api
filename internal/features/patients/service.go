package patients

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

type PatientUseCase struct {
	repo models.PatientRepository
	rd   *redis.Client
}

const patientCacheKeyPrefix = "patient:"

func NewPatientUseCase(repo models.PatientRepository, rd *redis.Client) models.PatientUseCase {
	return &PatientUseCase{repo: repo, rd: rd}
}

func (s *PatientUseCase) GetAll(ctx context.Context, page int, pageSize int) (*models.PaginatedPatients, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	cacheKey := fmt.Sprintf("%sall:%d:%d", patientCacheKeyPrefix, page, pageSize)

	cachedData, err := s.rd.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var result *models.PaginatedPatients
		if json.Unmarshal(cachedData, &result) == nil {
			return result, nil
		}
	} else if err != redis.Nil {
		fmt.Printf("Redis error: %v\n", err)
	}

	patients, total, err := s.repo.GetAll(offset, pageSize)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return nil, fmt.Errorf("no patients found")
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	if page > totalPages {
		return nil, fmt.Errorf("page %d exceeds total pages %d (total items: %d)", page, totalPages, total)
	}

	result := &models.PaginatedPatients{
		Data: patients,
		Paginate: models.Pagination{
			Total:      total,
			TotalPages: totalPages,
			Page:       page,
			PageSize:   pageSize,
		},
	}

	go func() {
		data, err := json.Marshal(result)
		if err != nil {
			fmt.Printf("Marshal error: %v\n", err)
			return
		}

		err = s.rd.Set(ctx, cacheKey, data, 5*time.Minute).Err()
		if err != nil {
			fmt.Printf("Redis set error: %v\n", err)
		}
	}()

	return result, nil
}

func (s *PatientUseCase) GetAllByVeterinarianID(ctx context.Context, veterinarianID uint64, page int, pageSize int) (*models.PaginatedPatients, error) {
	if veterinarianID == 0 {
		return nil, fmt.Errorf("veterinarian ID is required")
	}

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	cacheKey := fmt.Sprintf("%svet:%d:%d:%d", patientCacheKeyPrefix, veterinarianID, page, pageSize)

	cachedData, err := s.rd.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var result *models.PaginatedPatients
		if json.Unmarshal(cachedData, &result) == nil {
			return result, nil
		}
	} else if err != redis.Nil {
		fmt.Printf("Redis error: %v\n", err)
	}

	patients, total, err := s.repo.GetAllByVeterinarianID(veterinarianID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	if page > totalPages && total > 0 {
		return nil, fmt.Errorf("page %d exceeds total pages %d (total items: %d)", page, totalPages, total)
	}

	result := &models.PaginatedPatients{
		Data: patients,
		Paginate: models.Pagination{
			Total:      total,
			TotalPages: totalPages,
			Page:       page,
			PageSize:   pageSize,
		},
	}

	go func() {
		data, err := json.Marshal(result)
		if err != nil {
			fmt.Printf("Marshal error: %v\n", err)
			return
		}

		err = s.rd.Set(ctx, cacheKey, data, 5*time.Minute).Err()
		if err != nil {
			fmt.Printf("Redis set error: %v\n", err)
		}
	}()

	return result, nil
}

func (s *PatientUseCase) GetByID(id uint64) (*models.Patient, error) {
	if id == 0 {
		return nil, nil
	}

	patient, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return patient, nil
}

func (s *PatientUseCase) Create(ctx context.Context, p *models.Patient) error {
	newPatient, err := models.NewPatient(p.Name, p.Owner, p.OwnerEmail, p.OwnerPhone, p.Symptoms, p.VeterinarianID)

	if err != nil {
		return err
	}

	existingPatient, err := s.repo.GetByOwnerEmail(newPatient.OwnerEmail)
	if err == nil && existingPatient != nil {
		return fmt.Errorf("el correo electrónico ya se encuentra registrado")
	}

	err = s.repo.Create(newPatient)
	if err != nil {
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key value") {
			return fmt.Errorf("el correo electrónico ya se encuentra registrado")
		}
		return err
	}

	*p = *newPatient

	go func(vetID uint64) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pattern := fmt.Sprintf("%svet:%d:*", patientCacheKeyPrefix, vetID)
		iter := s.rd.Scan(bgCtx, 0, pattern, 0).Iterator()

		for iter.Next(bgCtx) {
			if err := s.rd.Del(bgCtx, iter.Val()).Err(); err != nil {
				log.Printf("Error invalidando caché %s: %v", iter.Val(), err)
			}
		}
	}(p.VeterinarianID)

	p = newPatient

	return nil
}

func (s *PatientUseCase) Update(ctx context.Context, id uint64, p *models.Patient) error {
	if id == 0 {
		return nil
	}

	log.Println("data", id, p)

	updatedPatient, err := models.NewPatient(p.Name, p.Owner, p.OwnerEmail, p.OwnerPhone, p.Symptoms, p.VeterinarianID)

	if err != nil {
		return err
	}

	err = s.repo.Update(id, updatedPatient)
	if err != nil {
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key value") {
			return fmt.Errorf("el correo electrónico ya se encuentra registrado")
		}
		return err
	}

	*p = *updatedPatient

	go func(vetID uint64) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pattern := fmt.Sprintf("%svet:%d:*", patientCacheKeyPrefix, vetID)
		iter := s.rd.Scan(bgCtx, 0, pattern, 0).Iterator()

		for iter.Next(bgCtx) {
			if err := s.rd.Del(bgCtx, iter.Val()).Err(); err != nil {
				log.Printf("Error invalidando caché %s: %v", iter.Val(), err)
			}
		}
	}(p.VeterinarianID)

	return nil
}

func (s *PatientUseCase) UpdateStatus(id uint64, status string) error {
	if id == 0 {
		return nil
	}

	if status == "" || (status != "active" && status != "inactive") {
		return fmt.Errorf("invalid status value: %s", status)
	}

	err := s.repo.UpdateStatus(id, status)
	if err != nil {
		return err
	}

	return nil
}

func (s *PatientUseCase) Delete(id uint64, veterinarianID uint64) error {
	if id == 0 {
		return nil
	}

	err := s.repo.Delete(id, veterinarianID)
	if err != nil {
		return err
	}

	return nil
}
