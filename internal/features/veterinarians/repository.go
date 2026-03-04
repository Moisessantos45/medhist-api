package veterinarians

import (
	"api_citas/internal/pkg/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) models.VeterinarianRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetAll(offset int, limit int) ([]models.Veterinarian, int64, error) {
	var veterinarians []models.Veterinarian
	var total int64

	err := r.db.Model(&models.Veterinarian{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Limit(limit).Offset(offset).Find(&veterinarians).Error
	if err != nil {
		return nil, 0, err
	}

	return veterinarians, total, nil
}

func (r *PostgresRepository) GetByID(id uint64) (*models.Veterinarian, error) {
	var veterinarian models.Veterinarian
	err := r.db.Where("email_confirmed = ?", true).First(&veterinarian, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("veterinario no encontrado con email confirmado")
		}
		return nil, err
	}
	return &veterinarian, nil
}

func (r *PostgresRepository) GetByEmail(name string, emailConfirmed bool) (*models.Veterinarian, error) {
	var veterinarian models.Veterinarian
	err := r.db.Where("email = ?", name).Where("email_confirmed = ?", emailConfirmed).First(&veterinarian).Error
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("veterinario no encontrado con email confirmado")
		}

		return nil, err
	}

	return &veterinarian, nil
}

func (r *PostgresRepository) Create(veterinarian *models.Veterinarian) error {
	return r.db.Create(veterinarian).Error
}

func (r *PostgresRepository) Update(id uint64, veterinarian *models.Veterinarian) error {
	var existingVeterinarian models.Veterinarian
	err := r.db.Where("email_confirmed = ?", true).First(&existingVeterinarian, id).Error
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("veterinario no encontrado con email confirmado")
		}

		return err
	}

	existingVeterinarian.Name = veterinarian.Name
	existingVeterinarian.Email = veterinarian.Email
	existingVeterinarian.Phone = veterinarian.Phone
	existingVeterinarian.Website = veterinarian.Website

	return r.db.Model(&models.Veterinarian{}).Where("id = ?", id).Updates(existingVeterinarian).Error
}

func (r *PostgresRepository) UpdatePassword(id uint64, newPassword string) error {
	return r.db.Model(&models.Veterinarian{}).Where("id = ?", id).Update("password", newPassword).Error
}

func (r *PostgresRepository) UpdateEmailConfirmed(id uint64, emailConfirmed bool) error {
	return r.db.Model(&models.Veterinarian{}).Where("id = ?", id).Update("email_confirmed", emailConfirmed).Error
}

func (r *PostgresRepository) UpdateToken(id uint64, token string) error {
	return r.db.Model(&models.Veterinarian{}).Where("id = ?", id).Update("token", token).Error
}

func (r *PostgresRepository) Delete(id uint64) error {
	return r.db.Delete(&models.Veterinarian{}, id).Error
}
