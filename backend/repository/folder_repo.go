package repository

import (
	"pipec-backend/models"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Folder = models.Folder

func GetFoldersByUser(db *gorm.DB, userID uint) ([]Folder, error) {
	var folders []Folder
	err := db.Where("user_id = ?", userID).Find(&folders).Error
	return folders, err
}

func GetOrCreateFolderByName(db *gorm.DB, userID uint, name string) (*Folder, error) {
	var folder Folder
	err := db.Where("user_id = ? AND name = ?", userID, name).First(&folder).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new folder if not found
			folder = Folder{
				UserID: userID,
				Name:   name,
			}
			err = db.Create(&folder).Error
			if err != nil {
				return nil, err
			}
			return &folder, nil
		}
		return nil, err
	}
	return &folder, nil
}

func CountMessagesInFolder(db *gorm.DB, folderID uint) (int64, error) {
	var count int64
	err := db.Model(&Message{}).Where("id = ?", folderID).Count(&count).Error
	return count, err
}

func CountRecentMessagesInFolder(db *gorm.DB, folderID uint, duration time.Duration) (int64, error) {
	var count int64
	since := time.Now().Add(-duration)
	err := db.Model(&Message{}).Where("id = ? AND created_at > ?", folderID, since).Count(&count).Error
	return count, err
}

func CountUnseenMessagesInFolder(db *gorm.DB, folderID uint) (int64, error) {
	var count int64

	err := db.Model(&Message{}).
		Where("id = ? AND NOT (flags @> ?)", folderID, pq.Array([]string{"SEEN"})).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}
