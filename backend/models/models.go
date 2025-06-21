package models

import (
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Mailbox struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        int            `json:"id"`
	MailboxID int            `json:"mailbox_id"`
	UID       string         `json:"uid"`
	FromAddr  string         `json:"from_addr"`
	ToAddrs   []string       `json:"to_addrs"`
	Subject   string         `json:"subject"`
	Body      string         `json:"body"`
	Flags     []string       `json:"flags"`
	Headers   map[string]any `json:"headers"`
	CreatedAt time.Time      `json:"created_at"`
}

type MailboxInfo struct {
	Name     string `json:"name"`
	Messages int    `json:"messages"`
	Recent   int    `json:"recent"`
	Unseen   int    `json:"unseen"`
}

// Protocol structures
type Command struct {
	ID      string                 `json:"id"`
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:"params"`
}

type Response struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
