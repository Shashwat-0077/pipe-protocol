package main

import (
	"log"
	"os"
	"pipec-backend/db"
	"pipec-backend/server"

	"golang.org/x/crypto/bcrypt"
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
	db, err := db.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create default user and mailboxes for testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	var userID int
	db.QueryRow(`INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) 
				 ON CONFLICT (username) DO UPDATE SET email=EXCLUDED.email 
				 RETURNING id`,
		"testuser", "test@pipec.local", string(hashedPassword)).Scan(&userID)

	// Create default mailboxes
	defaultMailboxes := []string{"INBOX", "SENT", "DRAFTS", "TRASH"}
	for _, mbox := range defaultMailboxes {
		db.Exec(`INSERT INTO mailboxes (user_id, name) VALUES ($1, $2) ON CONFLICT (user_id, name) DO NOTHING`,
			userID, mbox)
	}

	// Start server
	server := server.NewPipeCServer(db, port)
	log.Fatal(server.Start())
}
