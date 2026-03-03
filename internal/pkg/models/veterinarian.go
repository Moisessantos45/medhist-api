package models

import (
	"api_citas/internal/pkg"
	"context"
	"fmt"
	"strings"
	"time"
)

type Veterinarian struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"type:varchar(125)" json:"name"`
	Email          string    `gorm:"type:varchar(125);uniqueIndex" json:"email"`
	Password       string    `gorm:"type:varchar(64)" json:"password"`
	Phone          string    `gorm:"type:varchar(20)" json:"phone"`
	Website        string    `gorm:"type:varchar(255)" json:"website"`
	Token          string    `gorm:"type:text" json:"token"`
	EmailConfirmed bool      `gorm:"default:false" json:"email_confirmed"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoCreateTime;autoUpdateTime" json:"updated_at"`

	Patient       []Patient       `gorm:"foreignKey:VeterinarianID;references:ID"`
	Appointment   []Appointment   `gorm:"foreignKey:VeterinarianID;references:ID"`
	MedicalRecord []MedicalRecord `gorm:"foreignKey:VeterinarianID;references:ID"`
	Vaccination   []Vaccination   `gorm:"foreignKey:VeterinarianID;references:ID"`
}

type PaginatedVeterinarians struct {
	Data     []Veterinarian `json:"data"`
	Paginate Pagination     `json:"paginate"`
}

type VeterinarianRepository interface {
	GetAll(offset int, limit int) ([]Veterinarian, int64, error)
	GetByID(id uint64) (*Veterinarian, error)
	GetByEmail(email string) (*Veterinarian, error)
	Create(veterinarian *Veterinarian) error
	Update(id uint64, veterinarian *Veterinarian) error
	UpdatePassword(id uint64, newPassword string) error
	UpdateEmailConfirmed(id uint64, emailConfirmed bool) error
	UpdateToken(id uint64, token string) error
	Delete(id uint64) error
}

type VeterinarianUseCase interface {
	GetAll(ctx context.Context, page int, pageSize int) (*PaginatedVeterinarians, error)
	GetByID(id uint64) (*Veterinarian, error)
	GetByEmail(email string) (*Veterinarian, error)
	Create(ctx context.Context, veterinarian *Veterinarian) error
	Update(id uint64, veterinarian *Veterinarian) error
	ChangePassword(id uint64, currentPassword string, newPassword string) error
	ResetPassword(ctx context.Context, id uint64, token string, newPassword string) error
	UpdateEmailConfirmed(ctx context.Context, id uint64) error
	UpdateToken(id uint64, token string) error
	Delete(id uint64) error
}

func NewVeterinarian(name, email, password, phone, website string, checkPassword bool) (*Veterinarian, error) {

	if name == "" || len(strings.TrimSpace(name)) < 3 || len(strings.TrimSpace(name)) > 125 {
		return nil, fmt.Errorf("name is required")
	}

	if checkPassword && len(strings.TrimSpace(password)) < 7 {
		return nil, fmt.Errorf("password must be at least 7 characters long")
	}

	email = strings.TrimSpace(email)
	if email == "" || len(email) < 10 || len(email) > 125 || !pkg.EmailRegex.MatchString(email) {
		return nil, fmt.Errorf("Email is required, must be 10-125 characters and valid format")
	}

	phone = strings.TrimSpace(phone)
	if phone == "" || len(phone) < 7 || len(phone) > 20 || !pkg.PhoneRegex.MatchString(phone) {
		return nil, fmt.Errorf("Phone is required, must be 7-20 characters consisting of numbers, spaces, -, (, ) and optional + prefix")
	}

	return &Veterinarian{
		Name:     name,
		Email:    email,
		Password: password,
		Phone:    phone,
		Website:  website,
	}, nil
}
