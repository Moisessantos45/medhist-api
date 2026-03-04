package veterinarians

import (
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/models"
	"api_citas/internal/templates"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type VeterinarianUseCase struct {
	repo models.VeterinarianRepository
	rd   *redis.Client
	mk   *pkg.PasetoMaker
}

const veterinarianCacheKeyPrefix = "veterinarian:"

func NewVeterinarianUseCase(repo models.VeterinarianRepository, rd *redis.Client, mk *pkg.PasetoMaker) models.VeterinarianUseCase {
	return &VeterinarianUseCase{
		repo: repo,
		rd:   rd,
		mk:   mk,
	}
}

func (s *VeterinarianUseCase) GetAll(ctx context.Context, page int, pageSize int) (*models.PaginatedVeterinarians, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	cacheKey := fmt.Sprintf("%sall:%d:%d", veterinarianCacheKeyPrefix, page, pageSize)

	cachedData, err := s.rd.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var result *models.PaginatedVeterinarians
		if json.Unmarshal(cachedData, &result) == nil {
			return result, nil
		}
	} else if err != redis.Nil {
		log.Printf("Redis error: %v", err)
	}

	vets, total, err := s.repo.GetAll(offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if page > totalPages && total > 0 {
		return nil, fmt.Errorf("page %d exceeds %d", page, totalPages)
	}

	result := &models.PaginatedVeterinarians{
		Data: vets,
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
			log.Printf("Marshal error: %v", err)
			return
		}

		err = s.rd.Set(ctx, cacheKey, data, 5*time.Minute).Err()
		if err != nil {
			log.Printf("Cache set error: %v", err)
		}
	}()

	return result, nil
}

func (s *VeterinarianUseCase) GetByID(id uint64) (*models.Veterinarian, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid ID")
	}

	veterinarian, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	veterinarian.Password = ""

	return veterinarian, nil
}

func (s *VeterinarianUseCase) GetByEmail(email string) (*models.Veterinarian, error) {
	if email == "" || len(email) == 0 || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("email cannot be empty")
	}

	veterinarian, err := s.repo.GetByEmail(email, true)
	if err != nil {
		return nil, err
	}

	return veterinarian, nil
}

func (s *VeterinarianUseCase) Create(ctx context.Context, veterinarian *models.Veterinarian) error {
	isProduction := os.Getenv("GO_ENV")
	var host = os.Getenv("HOST_URL_PROD")
	if isProduction == "dev" {
		host = os.Getenv("HOST_URL_DEV")
	}

	newVeterinarian, err := models.NewVeterinarian(veterinarian.Name, veterinarian.Email, veterinarian.Password, veterinarian.Phone, veterinarian.Website, true)
	if err != nil {
		return err
	}

	existingVet, err := s.repo.GetByEmail(newVeterinarian.Email, true)
	if err == nil && existingVet != nil {
		return fmt.Errorf("email already in use")
	}

	hashedPassword, err := pkg.HashPassword(newVeterinarian.Password)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	newVeterinarian.Password = hashedPassword

	err = s.repo.Create(newVeterinarian)
	if err != nil {
		return err
	}

	token, err := s.mk.NewToken(fmt.Sprintf("%d", newVeterinarian.ID), 15*time.Minute)
	if err != nil {
		return fmt.Errorf("Error generating token")
	}

	newVeterinarian.Token = token

	err = s.rd.Set(ctx, token, newVeterinarian.ID, 15*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("error caching token: %w", err)
	}

	renderer, err := templates.NewEmailRenderer()
	if err != nil {
		return err
	}

	data := templates.ConfirmAccountData{
		Name:        newVeterinarian.Name,
		ConfirmLink: fmt.Sprintf("%s/confirm/%s", host, token),
	}

	htmlContent, err := renderer.RenderConfirmAccount(data)
	if err != nil {
		return err
	}

	err = pkg.SendEmail(ctx, []string{veterinarian.Email}, "Confirmacion de cuenta", htmlContent)
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}

func (s *VeterinarianUseCase) Update(id uint64, veterinarian *models.Veterinarian) error {
	if id == 0 {
		return fmt.Errorf("invalid ID")
	}

	newVeterinarian, err := models.NewVeterinarian(veterinarian.Name, veterinarian.Email, veterinarian.Password, veterinarian.Phone, veterinarian.Website, false)
	if err != nil {
		return err
	}

	err = s.repo.Update(id, newVeterinarian)
	if err != nil {
		return err
	}
	return nil
}

func (s *VeterinarianUseCase) ChangePassword(id uint64, currentPassword string, newPassword string) error {
	if id == 0 {
		return fmt.Errorf("invalid ID")
	}

	if currentPassword == "" || len(currentPassword) < 6 {
		return fmt.Errorf("current password cannot be empty")
	}

	if newPassword == "" || len(newPassword) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	vet, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("error fetching veterinarian: %w", err)
	}

	if !pkg.CheckPasswordHash(currentPassword, vet.Password) {
		return fmt.Errorf("current password is incorrect")
	}

	hashedPassword, err := pkg.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	err = s.repo.UpdatePassword(id, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

func (s *VeterinarianUseCase) ResetPassword(ctx context.Context, id uint64, token string, newPassword string) error {
	if id == 0 {
		return fmt.Errorf("invalid ID")
	}

	if newPassword == "" || len(newPassword) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	vet, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("error fetching veterinarian: %w", err)
	}

	if vet.Token != token {
		return fmt.Errorf("invalid token")
	}

	hashedPassword, err := pkg.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	err = s.repo.UpdatePassword(id, hashedPassword)
	if err != nil {
		return err
	}

	err = s.repo.UpdateToken(id, "")
	if err != nil {
		return err
	}

	return nil
}

func (s *VeterinarianUseCase) UpdateEmailConfirmed(ctx context.Context, id uint64) error {
	if id == 0 {
		return fmt.Errorf("invalid ID")
	}

	err := s.repo.UpdateEmailConfirmed(id, true)
	if err != nil {
		return err
	}

	return nil
}

func (s *VeterinarianUseCase) UpdateToken(id uint64, token string) error {
	if id == 0 {
		return fmt.Errorf("invalid ID")
	}

	err := s.repo.UpdateToken(id, token)
	if err != nil {
		return err
	}
	return nil
}

func (s *VeterinarianUseCase) Delete(id uint64) error {
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	return nil
}
