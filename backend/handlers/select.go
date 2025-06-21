package handlers

import (
	"pipec-backend/models"
	"pipec-backend/types"
	"time"

	"gorm.io/gorm"
)

func handleSelect(db *gorm.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State == types.StateNotAuthenticated {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Not authenticated"})
		return
	}

	mailbox, ok := cmd.Params["mailbox"].(string)
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Mailbox name required"})
		return
	}

	var mbox models.Mailbox
	err := db.Where("user_id = ? AND name = ?", client.UserID, mailbox).First(&mbox).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Mailbox not found"})
		} else {
			send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Database error"})
		}
		return
	}

	var messageCount int64
	var recentCount int64
	var unseenCount int64

	// Get message count
	if err := db.Model(&models.Message{}).Where("mailbox_id = ?", mbox.ID).Count(&messageCount).Error; err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to count messages"})
		return
	}

	// Get recent messages count (last 24 hours)
	if err := db.Model(&models.Message{}).Where("mailbox_id = ? AND created_at > ?", mbox.ID, time.Now().Add(-24*time.Hour)).Count(&recentCount).Error; err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to count recent messages"})
		return
	}

	// Get unseen messages count (messages without SEEN flag)
	// Try PostgreSQL JSON contains operator first
	err = db.Model(&models.Message{}).Where("mailbox_id = ? AND NOT (flags @> ?)", mbox.ID, `["SEEN"]`).Count(&unseenCount).Error
	if err != nil {
		// Fallback: try with JSON_CONTAINS (MySQL) or alternative approach
		err = db.Model(&models.Message{}).Where("mailbox_id = ? AND NOT JSON_CONTAINS(flags, ?)", mbox.ID, `"SEEN"`).Count(&unseenCount).Error
		if err != nil {
			// Final fallback: manual counting
			var messages []models.Message
			if err := db.Select("flags").Where("mailbox_id = ?", mbox.ID).Find(&messages).Error; err != nil {
				send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to count unseen messages"})
				return
			}

			// Count unseen messages manually
			unseenCount = 0
			for _, msg := range messages {
				seen := false
				for _, flag := range msg.Flags {
					if flag == "SEEN" {
						seen = true
						break
					}
				}
				if !seen {
					unseenCount++
				}
			}
		}
	}

	client.State = types.StateSelected
	client.SelectedMbox = mailbox
	client.SelectedMboxID = int(mbox.ID)

	info := models.MailboxInfo{
		Name:     mailbox,
		Messages: int(messageCount),
		Recent:   int(recentCount),
		Unseen:   int(unseenCount),
	}

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Mailbox selected", Data: info})
}
