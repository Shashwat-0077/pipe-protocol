package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type Message struct {
	ID        uint              `gorm:"primaryKey"`
	UserID    uint              `gorm:"index;not null"`
	FolderID  uint              `gorm:"index;not null"`
	From      string            `gorm:"not null"`
	To        pq.StringArray    `gorm:"type:text[]"`
	Cc        pq.StringArray    `gorm:"type:text[]"`
	Bcc       pq.StringArray    `gorm:"type:text[]"`
	Subject   string            `gorm:"type:text"`
	Body      string            `gorm:"type:text"`
	Headers   datatypes.JSONMap `gorm:"type:jsonb"`
	Flags     pq.StringArray    `gorm:"type:text[]"`
	UID       uint              `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Folder Folder `gorm:"foreignKey:FolderID;constraint:OnDelete:CASCADE"`
	User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
