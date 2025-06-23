package models

import (
	"pipec-backend/enums"
	"time"
)

type Folder struct {
	ID        uint             `gorm:"primaryKey"`
	UserID    uint             `gorm:"index;not null"`
	Name      string           `gorm:"not null"`
	Type      enums.FolderType `gorm:"type:text;index"`
	CreatedAt time.Time
	UpdatedAt time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type FolderInfo struct {
	Name     string `json:"name"`
	Messages int    `json:"messages"`
	Recent   int    `json:"recent"`
	Unseen   int    `json:"unseen"`
}
