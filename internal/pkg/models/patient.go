package models

import (
	"api_citas/internal/pkg"
	"context"
	"fmt"
	"strings"
	"time"
)

type Patient struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"type:varchar(125)" json:"name"`
	Owner          string    `gorm:"type:varchar(125)" json:"owner"`
	OwnerEmail     string    `gorm:"type:varchar(125);uniqueIndex" json:"owner_email"`
	OwnerPhone     string    `gorm:"type:varchar(20)" json:"owner_phone"`
	Symptoms       string    `gorm:"type:text" json:"symptoms"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	VeterinarianID uint64    `gorm:"not null;index" json:"veterinarian_id"`
	Status         string    `gorm:"type:varchar(20);default:'active'" json:"status"` // active, inactive

	Appointment   []Appointment   `gorm:"foreignKey:PatientID;references:ID"`
	MedicalRecord []MedicalRecord `gorm:"foreignKey:PatientID;references:ID"`
	Vaccination   []Vaccination   `gorm:"foreignKey:PatientID;references:ID"`
	Veterinarian  Veterinarian    `gorm:"foreignKey:VeterinarianID;references:ID"`
}

type MedicalRecord struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	VisitDate    time.Time `gorm:"not null;index" json:"visit_date"`
	Diagnosis    string    `gorm:"type:text" json:"diagnosis"`
	Treatment    string    `gorm:"type:text" json:"treatment"`
	Prescription string    `gorm:"type:text" json:"prescription"`
	WeightKg     float64   `gorm:"type:decimal(5,2)" json:"weight_kg"`
	TemperatureC float64   `gorm:"type:decimal(4,2)" json:"temperature_c"`
	Notes        string    `gorm:"type:text" json:"notes"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	PatientID      uint64 `gorm:"not null;index" json:"patient_id"`
	VeterinarianID uint64 `gorm:"not null;index" json:"veterinarian_id"`

	Patient      Patient      `gorm:"foreignKey:PatientID;references:ID"`
	Veterinarian Veterinarian `gorm:"foreignKey:VeterinarianID;references:ID"`
}

type PaginatedPatients struct {
	Data     []Patient  `json:"data"`
	Paginate Pagination `json:"paginate"`
}

type PaginatedMedicalRecords struct {
	Data     []MedicalRecord `json:"data"`
	Paginate Pagination      `json:"paginate"`
}

type PatientRepository interface {
	GetAll(offset int, limit int) ([]Patient, int64, error)
	GetAllByVeterinarianID(veterinarianID uint64, offset int, limit int) ([]Patient, int64, error)
	GetByID(id uint64) (*Patient, error)
	GetByOwnerEmail(ownerEmail string) (*Patient, error)
	Create(patient *Patient) error
	Update(id uint64, patient *Patient) error
	UpdateStatus(id uint64, status string) error
	Delete(id uint64, veterinarianID uint64) error
}

type PatientUseCase interface {
	GetAll(ctx context.Context, page int, pageSize int) (*PaginatedPatients, error)
	GetAllByVeterinarianID(ctx context.Context, veterinarianID uint64, page int, pageSize int) (*PaginatedPatients, error)
	GetByID(id uint64) (*Patient, error)
	Create(ctx context.Context, patient *Patient) error
	Update(ctx context.Context, id uint64, patient *Patient) error
	UpdateStatus(id uint64, status string) error
	Delete(id uint64, veterinarianID uint64) error
}

type MedicalRecordRepository interface {
	GetAll(id uint64, veterinarianId uint64, offset int, limit int) ([]MedicalRecord, int64, error)
	GetByID(id uint64) (*MedicalRecord, error)
	Create(mr *MedicalRecord) error
	Update(id uint64, mr *MedicalRecord) error
	Delete(id uint64, veterinarianID uint64) error
}

type MedicalRecordUseCase interface {
	GetAll(ctx context.Context, id uint64, veterinarianId uint64, page int, pageSize int) (*PaginatedMedicalRecords, error)
	GetByID(id uint64) (*MedicalRecord, error)
	Create(ctx context.Context, mr *MedicalRecord) error
	Update(ctx context.Context, id uint64, mr *MedicalRecord) error
	Delete(id uint64, veterinarianID uint64) error
}

func NewPatient(name string, owner string, ownerEmail string, ownerPhone string, symptoms string, VeterinarianID uint64) (*Patient, error) {

	if VeterinarianID == 0 {
		return nil, fmt.Errorf("veterinarian id is required")
	}

	name = strings.TrimSpace(name)
	if name == "" || len(name) < 3 || len(name) > 125 {
		return nil, fmt.Errorf("name is required and must be 3-125 characters")
	}

	owner = strings.TrimSpace(owner)
	if owner == "" || len(owner) < 3 || len(owner) > 125 {
		return nil, fmt.Errorf("owner is required and must be 3-125 characters")
	}

	ownerEmail = strings.TrimSpace(ownerEmail)
	if ownerEmail == "" || len(ownerEmail) < 10 || len(ownerEmail) > 125 || !pkg.EmailRegex.MatchString(ownerEmail) {
		return nil, fmt.Errorf("owner email is required, must be 10-125 characters and valid format")
	}

	ownerPhone = strings.TrimSpace(ownerPhone)
	if ownerPhone == "" || len(ownerPhone) < 7 || len(ownerPhone) > 20 || !pkg.PhoneRegex.MatchString(ownerPhone) {
		return nil, fmt.Errorf("owner phone is required, must be 7-20 characters consisting of numbers, spaces, -, (, ) and optional + prefix")
	}

	return &Patient{
		Name:           name,
		Owner:          owner,
		OwnerEmail:     ownerEmail,
		OwnerPhone:     ownerPhone,
		Symptoms:       symptoms,
		VeterinarianID: VeterinarianID,
	}, nil
}

func NewMedicalRecord(visitDate time.Time, diagnosis string, treatment string, prescription string, weightKg, temperatureC float64, notes string, patientId uint64, veterinarianId uint64) (*MedicalRecord, error) {

	if patientId == 0 {
		return nil, fmt.Errorf("patient id is required")
	}

	if veterinarianId == 0 {
		return nil, fmt.Errorf("veterinarian id is required")
	}

	if visitDate.IsZero() {
		return nil, fmt.Errorf("visit date is required")
	}

	if diagnosis == "" {
		return nil, fmt.Errorf("diagnosis is required")
	}

	if treatment == "" {
		return nil, fmt.Errorf("treatment is required")
	}

	if prescription == "" {
		return nil, fmt.Errorf("prescription is required")
	}

	if weightKg <= 0 {
		return nil, fmt.Errorf("weight must be greater than 0")
	}

	if temperatureC <= 0 {
		return nil, fmt.Errorf("temperature must be greater than 0")
	}

	return &MedicalRecord{
		VisitDate:      visitDate,
		Diagnosis:      diagnosis,
		Treatment:      treatment,
		Prescription:   prescription,
		WeightKg:       weightKg,
		TemperatureC:   temperatureC,
		Notes:          notes,
		PatientID:      patientId,
		VeterinarianID: veterinarianId,
	}, nil
}
