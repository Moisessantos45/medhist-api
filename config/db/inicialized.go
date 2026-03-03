package db

import (
	"api_citas/internal/pkg/models"
	"fmt"
)

func Inicialized() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := DB.AutoMigrate(models.Models...); err != nil {
		return fmt.Errorf("error migrating database: %v", err)
	}

	return nil
}
