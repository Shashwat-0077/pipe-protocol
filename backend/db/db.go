package db

import (
	"pipec-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database initialization
func InitDB(dbURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Mailbox{},
		&models.Message{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
