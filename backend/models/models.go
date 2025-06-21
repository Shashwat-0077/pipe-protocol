package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Custom JSON type for GORM
type JSON []string

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSON", value)
	}

	return json.Unmarshal(bytes, j)
}

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Custom JSON type for headers (map)
type JSONMap map[string]interface{}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	return json.Unmarshal(bytes, j)
}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships - HasMany relationship
	Mailboxes []Mailbox `json:"mailboxes,omitempty" gorm:"foreignKey:UserID"`
}

type Mailbox struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index;constraint:OnDelete:CASCADE"`
	Name      string    `json:"name" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships - BelongsTo relationship
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Messages []Message `json:"messages,omitempty" gorm:"foreignKey:MailboxID"`
}

// Add unique constraint for user_id and name combination
func (Mailbox) TableName() string {
	return "mailboxes"
}

// Add unique index in AfterAutoMigrate hook or use gorm tags
type Message struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MailboxID uint      `json:"mailbox_id" gorm:"not null;index;constraint:OnDelete:CASCADE"`
	UID       string    `json:"uid" gorm:"uniqueIndex;not null"`
	FromAddr  string    `json:"from_addr" gorm:"not null"`
	ToAddrs   JSON      `json:"to_addrs" gorm:"type:jsonb;not null"`
	Subject   string    `json:"subject" gorm:"type:text"`
	Body      string    `json:"body" gorm:"type:text"`
	Flags     JSON      `json:"flags" gorm:"type:jsonb;default:'[]'"`
	Headers   JSONMap   `json:"headers" gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships - BelongsTo relationship
	Mailbox Mailbox `json:"mailbox,omitempty" gorm:"foreignKey:MailboxID"`
}

type MailboxInfo struct {
	Name     string `json:"name"`
	Messages int    `json:"messages"`
	Recent   int    `json:"recent"`
	Unseen   int    `json:"unseen"`
}

// Protocol structures (unchanged)
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
