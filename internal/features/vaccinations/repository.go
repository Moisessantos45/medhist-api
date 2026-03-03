package vaccinations

import (
	"api_citas/internal/pkg/models"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) models.VaccinationRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetAll(offset int, limit int) ([]models.Vaccination, int64, error) {
	var vaccinations []models.Vaccination
	var total int64

	err := r.db.Model(&models.Vaccination{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Limit(limit).Offset(offset).Find(&vaccinations).Error
	if err != nil {
		return nil, 0, err
	}

	return vaccinations, total, nil
}

func (r *PostgresRepository) GetByID(id uint64) (*models.Vaccination, error) {
	var vaccination models.Vaccination
	err := r.db.First(&vaccination, id).Error

	if err != nil {
		return nil, err
	}

	return &vaccination, nil
}

func (r *PostgresRepository) Create(vaccination *models.Vaccination) error {
	return r.db.Create(vaccination).Error
}

func (r *PostgresRepository) Update(id uint64, vaccination *models.Vaccination) error {
	var existingVaccination models.Vaccination
	err := r.db.First(&existingVaccination, id).Error
	if err != nil {
		return err
	}

	existingVaccination.Date = vaccination.Date
	existingVaccination.Type = vaccination.Type
	existingVaccination.NextDueDate = vaccination.NextDueDate
	existingVaccination.PatientID = vaccination.PatientID
	existingVaccination.VeterinarianID = vaccination.VeterinarianID

	return r.db.Save(&existingVaccination).Error
}

func (r *PostgresRepository) UpdateStatus(id uint64, status string) error {
	return r.db.Model(&models.Vaccination{}).Where("id = ?", id).Update("status", status).Error
}

func (r *PostgresRepository) Delete(id uint64, veterinarianID uint64) error {
	return r.db.Where("id = ? AND veterinarian_id = ?", id, veterinarianID).Delete(&models.Vaccination{}).Error
}
