package utils

import (
	"pipec-backend/models"

	"gorm.io/gorm"
)

func GenerateUID(db *gorm.DB, userID uint) uint {
	var maxUID uint
	db.Model(&models.Message{}).Where("user_id = ?", userID).Select("COALESCE(MAX(uid), 0)").Scan(&maxUID)
	return maxUID + 1
}
