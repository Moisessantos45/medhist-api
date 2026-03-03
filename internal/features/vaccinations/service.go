package vaccinations

import (
	"api_citas/internal/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

type VaccinationUseCase struct {
	repo models.VaccinationRepository
	rd   *redis.Client
}

const vaccinationCacheKeyPrefix = "vaccination:"

func NewVaccinationUseCase(repo models.VaccinationRepository, rd *redis.Client) models.VaccinationUseCase {
	return &VaccinationUseCase{repo: repo, rd: rd}
}

func (s *VaccinationUseCase) GetAll(ctx context.Context, patientID uint64, veterinarianID uint64, page int, pageSize int) (*models.PaginatedVaccinations, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	cacheKey := fmt.Sprintf("%sall:%d:%d:%d:%d", vaccinationCacheKeyPrefix, patientID, veterinarianID, page, pageSize)

	cachedData, err := s.rd.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var result *models.PaginatedVaccinations
		if err := json.Unmarshal(cachedData, &result); err == nil {
			return result, nil
		}
	} else if err != redis.Nil {
		fmt.Printf("Redis error: %v\n", err)
	}

	vaccinations, total, err := s.repo.GetAll(offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	if page > totalPages && total > 0 {
		return nil, fmt.Errorf("page %d exceeds total pages %d (total items: %d)", page, totalPages, total)
	}

	result := &models.PaginatedVaccinations{
		Data: vaccinations,
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

		if err := s.rd.Set(ctx, cacheKey, data, 0).Err(); err != nil {
			fmt.Printf("Redis set error: %v\n", err)
		}
	}()

	return result, nil
}

func (s *VaccinationUseCase) GetByID(id uint64) (*models.Vaccination, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid ID: %d", id)
	}

	return s.repo.GetByID(id)
}

func (s *VaccinationUseCase) Create(ctx context.Context, v *models.Vaccination) error {
	newVaccination, err := models.NewVaccination(v.Type, v.Date, v.NextDueDate, v.PatientID, v.VeterinarianID)
	if err != nil {
		return fmt.Errorf("error creating vaccination: %w", err)
	}

	err = s.repo.Create(newVaccination)
	if err != nil {
		return fmt.Errorf("error creating vaccination: %w", err)
	}

	*v = *newVaccination

	go func(patientID uint64, veterinarianID uint64) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pattern := fmt.Sprintf("%sall:%d:%d:*", vaccinationCacheKeyPrefix, patientID, veterinarianID)
		iter := s.rd.Scan(bgCtx, 0, pattern, 0).Iterator()

		for iter.Next(bgCtx) {
			if err := s.rd.Del(bgCtx, iter.Val()).Err(); err != nil {
				fmt.Printf("Error deleting cache key %s: %v\n", iter.Val(), err)
			}
		}
	}(v.PatientID, v.VeterinarianID)

	return nil
}

func (s *VaccinationUseCase) Update(ctx context.Context, id uint64, v *models.Vaccination) error {
	if id == 0 {
		return fmt.Errorf("invalid ID: %d", id)
	}

	newVaccination, err := models.NewVaccination(v.Type, v.Date, v.NextDueDate, v.PatientID, v.VeterinarianID)
	if err != nil {
		return fmt.Errorf("error creating vaccination: %w", err)
	}

	err = s.repo.Update(id, newVaccination)
	if err != nil {
		return fmt.Errorf("error updating vaccination: %w", err)
	}

	*v = *newVaccination

	go func(patientID uint64, veterinarianID uint64) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pattern := fmt.Sprintf("%sall:%d:%d:*", vaccinationCacheKeyPrefix, patientID, veterinarianID)
		iter := s.rd.Scan(bgCtx, 0, pattern, 0).Iterator()

		for iter.Next(bgCtx) {
			if err := s.rd.Del(bgCtx, iter.Val()).Err(); err != nil {
				fmt.Printf("Error deleting cache key %s: %v\n", iter.Val(), err)
			}
		}
	}(v.PatientID, v.VeterinarianID)

	return nil
}

func (s *VaccinationUseCase) UpdateStatus(id uint64, status string) error {
	if id == 0 {
		return fmt.Errorf("invalid ID: %d", id)
	}

	if status != "completed" && status != "pending" && status != "canceled" {
		return fmt.Errorf("invalid status: %s", status)
	}

	return s.repo.UpdateStatus(id, status)
}

func (s *VaccinationUseCase) Delete(id uint64, veterinarianID uint64) error {
	if id == 0 {
		return fmt.Errorf("invalid ID: %d", id)
	}
	return s.repo.Delete(id, veterinarianID)
}
