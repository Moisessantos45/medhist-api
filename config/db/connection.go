package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() error {
	var err error
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	user := os.Getenv("DBUSER")
	password := os.Getenv("PASSWORD")
	dbname := os.Getenv("DBNAME")

	DNS := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	DB, err = gorm.Open(postgres.Open(DNS), &gorm.Config{})
	if err != nil {
		return err
	}

	return nil
}
