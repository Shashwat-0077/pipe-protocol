package main

import (
	"log"
	"os"
	"pipec-backend/config"
	"pipec-backend/models"
	"pipec-backend/server"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:shasha@localhost/piper?sslmode=disable"
	}

	port := os.Getenv("PIPEC_PORT")
	if port == "" {
		port = "1143"
	}

	// Initialize database
	database, err := config.ConnectDatabase(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Get underlying SQL DB for closing connection
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying SQL DB: %v", err)
	}
	defer sqlDB.Close()

	// Create default user and folder for testing
	err = createDefaultData(database)
	if err != nil {
		log.Printf("Warning: Failed to create default data: %v", err)
	}

	// Start server
	server := server.NewPipeCServer(database, port)
	log.Fatal(server.Start())
}

func createDefaultData(db *gorm.DB) error {
	// Create default user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		Username: "testuser",
		Email:    "test|pipec.local",
		Password: string(hashedPassword),
	}

	// Use FirstOrCreate to avoid duplicates
	err = db.Where("username = ?", user.Username).FirstOrCreate(&user).Error
	if err != nil {
		return err
	}

	// Create default folder
	defaultFolders := []string{"INBOX", "SENT", "DRAFTS", "TRASH"}
	for _, mboxName := range defaultFolders {
		folder := models.Folder{
			UserID: user.ID,
			Name:   mboxName,
		}

		// Use FirstOrCreate to avoid duplicates
		err = db.Where("user_id = ? AND name = ?", user.ID, mboxName).FirstOrCreate(&folder).Error
		if err != nil {
			return err
		}
	}

	log.Printf("Default user 'testuser' created/updated with folders: %v", defaultFolders)
	return nil
}
