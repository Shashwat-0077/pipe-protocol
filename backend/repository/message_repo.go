package repository

import (
	"pipec-backend/models"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Message = models.Message

func ListMessagesByFolder(db *gorm.DB, folderID uint) ([]Message, error) {
	var msgs []Message
	err := db.Where("folder_id = ?", folderID).Find(&msgs).Error
	return msgs, err
}

func GetMessageByID(db *gorm.DB, id uint) (*Message, error) {
	var msg Message
	err := db.First(&msg, id).Error
	return &msg, err
}

func UpdateMessageFlags(db *gorm.DB, id uint, flags []string) error {
	return db.Model(&Message{}).Where("id = ?", id).Update("flags", pq.StringArray(flags)).Error
}

func MoveMessageToFolder(db *gorm.DB, id uint, newFolderID uint) error {
	return db.Model(&Message{}).Where("id = ?", id).Update("folder_id", newFolderID).Error
}

func GetMessagesBySequenceRange(db *gorm.DB, folderID uint, start int, end int) ([]Message, error) {
	var messages []Message
	err := db.Where("folder_id = ?", folderID).Order("id").Offset(start - 1).Limit(end - start + 1).Find(&messages).Error
	return messages, err
}

func GetMessageBySequence(db *gorm.DB, folderID uint, seq int) (*Message, error) {
	var msg Message
	err := db.Where("folder_id = ?", folderID).Order("id").Offset(seq - 1).Limit(1).First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func CreateMessage(db *gorm.DB, message *Message) error {
	return db.Create(message).Error
}
