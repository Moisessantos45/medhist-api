package models

import (
	"context"
	"fmt"
	"time"
)

type Appointment struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Date      time.Time `gorm:"not null;index" json:"date"`
	Reason    string    `gorm:"type:text" json:"reason"`
	Status    string    `gorm:"type:varchar(20);default:'scheduled'" json:"status"` // scheduled, completed, canceled
	Notes     string    `gorm:"type:text" json:"notes"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	PatientID      uint64 `gorm:"not null;index" json:"patient_id"`
	VeterinarianID uint64 `gorm:"not null;index" json:"veterinarian_id"`

	Patient      Patient      `gorm:"foreignKey:PatientID;references:ID"`
	Veterinarian Veterinarian `gorm:"foreignKey:VeterinarianID;references:ID"`
}

type PaginatedAppointments struct {
	Data     []Appointment `json:"data"`
	Paginate Pagination    `json:"paginate"`
}

type AppointmentRepository interface {
	GetAll(patientID uint64, veterinarianID uint64, offset int, limit int) ([]Appointment, int64, error)
	GetByID(id uint64) (*Appointment, error)
	Create(appointment *Appointment) error
	Update(id uint64, appointment *Appointment) error
	Delete(id uint64, veterinarianID uint64) error
}

type AppointmentUseCase interface {
	GetAll(ctx context.Context, patientID uint64, veterinarianID uint64, page int, pageSize int) (*PaginatedAppointments, error)
	GetByID(id uint64) (*Appointment, error)
	Create(ctx context.Context, appointment *Appointment) error
	Update(ctx context.Context, id uint64, appointment *Appointment) error
	Delete(id uint64, veterinarianID uint64) error
}

func NewAppointment(date time.Time, reason string, status string, notes string, patientID uint64, veterinarianID uint64) (*Appointment, error) {

	if patientID == 0 {
		return nil, fmt.Errorf("patientID cannot be zero")
	}

	if veterinarianID == 0 {
		return nil, fmt.Errorf("veterinarianID cannot be zero")
	}

	if date.IsZero() {
		return nil, fmt.Errorf("date cannot be zero")
	}

	if status != "scheduled" && status != "completed" && status != "canceled" {
		status = "scheduled"
	}

	if reason == "" {
		reason = "General Checkup"
	}

	return &Appointment{
		Date:           date,
		Reason:         reason,
		Status:         status,
		Notes:          notes,
		PatientID:      patientID,
		VeterinarianID: veterinarianID,
	}, nil
}
