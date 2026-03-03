package medicalrecords

import (
	"api_citas/internal/pkg/models"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) models.MedicalRecordRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetAll(id uint64, veterinarianId uint64, offset int, limit int) ([]models.MedicalRecord, int64, error) {
	var medicalRecords []models.MedicalRecord
	var total int64

	if err := r.db.Model(&models.MedicalRecord{}).Where("patient_id = ? AND veterinarian_id = ?", id, veterinarianId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Where("patient_id = ? AND veterinarian_id = ?", id, veterinarianId).Limit(limit).Offset(offset).Find(&medicalRecords).Error; err != nil {
		return nil, 0, err
	}

	return medicalRecords, total, nil
}

func (r *PostgresRepository) GetByID(id uint64) (*models.MedicalRecord, error) {
	var medicalRecord models.MedicalRecord

	if err := r.db.First(&medicalRecord, id).Error; err != nil {
		return nil, err
	}

	return &medicalRecord, nil
}

func (r *PostgresRepository) Create(mr *models.MedicalRecord) error {
	if err := r.db.Create(mr).Error; err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) Update(id uint64, mr *models.MedicalRecord) error {
	var existingMedicalRecord models.MedicalRecord

	if err := r.db.First(&existingMedicalRecord, id).Error; err != nil {
		return err
	}

	existingMedicalRecord.VisitDate = mr.VisitDate
	existingMedicalRecord.Diagnosis = mr.Diagnosis
	existingMedicalRecord.Treatment = mr.Treatment
	existingMedicalRecord.Prescription = mr.Prescription
	existingMedicalRecord.WeightKg = mr.WeightKg
	existingMedicalRecord.TemperatureC = mr.TemperatureC
	existingMedicalRecord.Notes = mr.Notes

	// existingMedicalRecord.PatientID = mr.PatientID
	// existingMedicalRecord.VeterinarianID = mr.VeterinarianID

	if err := r.db.Save(&existingMedicalRecord).Error; err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) Delete(id uint64, veterinarianID uint64) error {

	if err := r.db.Where("id = ? AND veterinarian_id = ?", id, veterinarianID).Delete(&models.MedicalRecord{}).Error; err != nil {
		return err
	}

	return nil
}
