package models

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Vaccination struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"type:varchar(100)" json:"type"`
	Date        time.Time `gorm:"not null;index" json:"date"`
	NextDueDate time.Time `gorm:"not null;index" json:"next_due_date"`
	Status      string    `gorm:"type:varchar(20);default:'completed'" json:"status"` // completed, pending, canceled
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	PatientID      uint64 `gorm:"not null;index" json:"patient_id"`
	VeterinarianID uint64 `gorm:"not null;index" json:"veterinarian_id"`

	Patient      Patient      `gorm:"foreignKey:PatientID;references:ID"`
	Veterinarian Veterinarian `gorm:"foreignKey:VeterinarianID;references:ID"`
}

type PaginatedVaccinations struct {
	Data     []Vaccination `json:"data"`
	Paginate Pagination    `json:"paginate"`
}

type VaccinationRepository interface {
	GetAll(offset int, limit int) ([]Vaccination, int64, error)
	GetByID(id uint64) (*Vaccination, error)
	Create(vaccination *Vaccination) error
	Update(id uint64, vaccination *Vaccination) error
	UpdateStatus(id uint64, status string) error
	Delete(id uint64, veterinarianID uint64) error
}

type VaccinationUseCase interface {
	GetAll(ctx context.Context, patientID uint64, veterinarianID uint64, page int, pageSize int) (*PaginatedVaccinations, error)
	GetByID(id uint64) (*Vaccination, error)
	Create(ctx context.Context, vaccination *Vaccination) error
	Update(ctx context.Context, id uint64, vaccination *Vaccination) error
	UpdateStatus(id uint64, status string) error
	Delete(id uint64, veterinarianID uint64) error
}

func NewVaccination(vaccinationType string, date time.Time, nextDueDate time.Time, patientID uint64, veterinarianID uint64) (*Vaccination, error) {

	if date.IsZero() {
		return nil, fmt.Errorf("date is required")
	}

	if vaccinationType == "" || strings.TrimSpace(vaccinationType) == "" {
		return nil, fmt.Errorf("vaccination type is required")
	}

	if patientID == 0 {
		return nil, fmt.Errorf("patientID cannot be zero")
	}

	if veterinarianID == 0 {
		return nil, fmt.Errorf("veterinarianID cannot be zero")
	}

	return &Vaccination{
		Type:           vaccinationType,
		Date:           date,
		NextDueDate:    nextDueDate,
		PatientID:      patientID,
		VeterinarianID: veterinarianID,
	}, nil
}
