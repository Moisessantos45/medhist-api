package appointments

import (
	"api_citas/internal/pkg/models"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) models.AppointmentRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) GetAll(patientID uint64, veterinarianID uint64, offset int, limit int) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var total int64

	err := r.db.Model(&models.Appointment{}).Where("patient_id = ? AND veterinarian_id = ?", patientID, veterinarianID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("patient_id = ? AND veterinarian_id = ?", patientID, veterinarianID).Limit(limit).Offset(offset).Find(&appointments).Error
	if err != nil {
		return nil, 0, err
	}

	return appointments, total, nil
}

func (r *PostgresRepository) GetByID(id uint64) (*models.Appointment, error) {
	var appointment models.Appointment

	err := r.db.First(&appointment, id).Error
	if err != nil {
		return nil, err
	}

	return &appointment, nil
}

func (r *PostgresRepository) Create(appointment *models.Appointment) error {
	err := r.db.Create(appointment).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) Update(id uint64, a *models.Appointment) error {
	var existingAppointment models.Appointment

	err := r.db.First(&existingAppointment, id).Error
	if err != nil {
		return err
	}

	existingAppointment.Date = a.Date
	existingAppointment.Status = a.Status
	existingAppointment.Notes = a.Notes
	existingAppointment.PatientID = a.PatientID
	existingAppointment.VeterinarianID = a.VeterinarianID

	err = r.db.Save(&existingAppointment).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) Delete(id uint64, veterinarianID uint64) error {
	err := r.db.Where("id = ? AND veterinarian_id = ?", id, veterinarianID).Delete(&models.Appointment{}).Error
	if err != nil {
		return err
	}

	return nil
}
