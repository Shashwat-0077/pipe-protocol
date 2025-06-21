package db

import (
	"database/sql"

	_ "github.com/lib/pq" // this registers the "postgres" driver
)

// Database initialization
func InitDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS mailboxes (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, name)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		mailbox_id INTEGER REFERENCES mailboxes(id) ON DELETE CASCADE,
		uid VARCHAR(255) UNIQUE NOT NULL,
		from_addr VARCHAR(255) NOT NULL,
		to_addrs JSONB NOT NULL,
		subject TEXT,
		body TEXT,
		flags JSONB DEFAULT '[]',
		headers JSONB DEFAULT '{}',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_messages_mailbox_id ON messages(mailbox_id);
	CREATE INDEX IF NOT EXISTS idx_messages_uid ON messages(uid);
	`

	_, err = db.Exec(schema)
	return db, err
}
