package db

import (
	"api_citas/internal/pkg/models"
	"fmt"
	"os"
)

func Inicialized() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	prod := os.Getenv("GO_ENV")
	if prod == "production" {
		fmt.Println("Running in production mode, skipping database migration")
		return nil
	}

	if err := DB.AutoMigrate(models.Models...); err != nil {
		return fmt.Errorf("error migrating database: %v", err)
	}

	return nil
}
