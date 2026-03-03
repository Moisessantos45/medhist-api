package patients

import (
	"api_citas/internal/pkg/models"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) models.PatientRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetAll(offset int, limit int) ([]models.Patient, int64, error) {
	var patients []models.Patient
	var total int64

	err := r.db.Model(&models.Patient{}).Where("status = ?", "active").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("status = ?", "active").Limit(limit).Offset(offset).Find(&patients).Error
	if err != nil {
		return nil, 0, err
	}

	return patients, total, nil
}

func (r *PostgresRepository) GetAllByVeterinarianID(veterinarianID uint64, offset int, limit int) ([]models.Patient, int64, error) {
	var patients []models.Patient
	var total int64

	err := r.db.Model(&models.Patient{}).Where("veterinarian_id = ?", veterinarianID).Where("status = ?", "active").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("veterinarian_id = ?", veterinarianID).Where("status = ?", "active").Limit(limit).Offset(offset).Find(&patients).Error
	if err != nil {
		return nil, 0, err
	}

	return patients, total, nil
}

func (r *PostgresRepository) GetByID(id uint64) (*models.Patient, error) {
	var patient models.Patient
	err := r.db.First(&patient, id).Error
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *PostgresRepository) Create(patient *models.Patient) error {
	return r.db.Create(patient).Error
}

func (r *PostgresRepository) Update(id uint64, patient *models.Patient) error {
	var existingPatient models.Patient
	err := r.db.First(&existingPatient, id).Error
	if err != nil {
		return err
	}

	existingPatient.Name = patient.Name
	existingPatient.Owner = patient.Owner
	existingPatient.OwnerEmail = patient.OwnerEmail
	existingPatient.OwnerPhone = patient.OwnerPhone
	existingPatient.Symptoms = patient.Symptoms

	return r.db.Model(&models.Patient{}).Where("id = ?", id).Updates(existingPatient).Error
}

func (r *PostgresRepository) UpdateStatus(id uint64, status string) error {
	return r.db.Model(&models.Patient{}).Where("id = ?", id).Update("status", status).Error
}

func (r *PostgresRepository) Delete(id uint64, veterinarianID uint64) error {
	return r.db.Where("id = ? AND veterinarian_id = ?", id, veterinarianID).Delete(&models.Patient{}).Error
}
