package handlers

import (
	"pipec-backend/models"
	"pipec-backend/types"

	"gorm.io/gorm"
)

func handleList(db *gorm.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State == types.StateNotAuthenticated {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Not authenticated"})
		return
	}

	var mailboxes []models.Mailbox
	// Use Preload to load the associated User data
	err := db.Preload("User").Where("user_id = ?", client.UserID).Find(&mailboxes).Error
	if err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to list mailboxes"})
		return
	}

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Mailboxes listed", Data: mailboxes})
}
