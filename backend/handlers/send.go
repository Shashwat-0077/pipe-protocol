package handlers

import (
	"fmt"
	"pipec-backend/models"
	"pipec-backend/types"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func handleSend(db *gorm.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State == types.StateNotAuthenticated {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Not authenticated"})
		return
	}

	from, ok := cmd.Params["from"].(string)
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "From address required"})
		return
	}

	toInterface, ok := cmd.Params["to"].([]interface{})
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "To addresses required"})
		return
	}

	var to []string
	for _, addr := range toInterface {
		if str, ok := addr.(string); ok {
			to = append(to, str)
		}
	}

	subject, _ := cmd.Params["subject"].(string)
	body, _ := cmd.Params["body"].(string)

	// Find or create SENT mailbox
	var sentMbox models.Mailbox
	err := db.Where("user_id = ? AND name = ?", client.UserID, "SENT").First(&sentMbox).Error
	if err != nil {
		// Create SENT mailbox if it doesn't exist
		sentMbox = models.Mailbox{
			UserID: uint(client.UserID),
			Name:   "SENT",
		}
		err = db.Create(&sentMbox).Error
		if err != nil {
			send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to create SENT mailbox"})
			return
		}
	}

	uid := uuid.New().String()
	headers := models.JSONMap{
		"message-id": fmt.Sprintf("<%s@%s>", uid, "pipec.local"),
		"date":       time.Now().Format(time.RFC3339),
	}

	message := models.Message{
		MailboxID: sentMbox.ID,
		UID:       uid,
		FromAddr:  from,
		ToAddrs:   models.JSON(to),
		Subject:   subject,
		Body:      body,
		Flags:     models.JSON([]string{"SENT"}),
		Headers:   headers,
	}

	err = db.Create(&message).Error
	if err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to store message"})
		return
	}

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Message sent and stored"})
}
