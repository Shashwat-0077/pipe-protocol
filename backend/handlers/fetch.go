package handlers

import (
	"pipec-backend/models"
	"pipec-backend/types"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func handleFetch(db *gorm.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State != types.StateSelected {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "No mailbox selected"})
		return
	}

	sequence, ok := cmd.Params["sequence"].(string)
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Sequence required"})
		return
	}

	_, ok = cmd.Params["items"].([]interface{})
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Items required"})
		return
	}

	var messages []models.Message
	var err error

	if strings.Contains(sequence, ":") {
		// Range query
		parts := strings.Split(sequence, ":")
		if len(parts) == 2 {
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			limit := end - start + 1
			offset := start - 1

			err = db.Where("mailbox_id = ?", client.SelectedMboxID).
				Order("id").
				Limit(limit).
				Offset(offset).
				Find(&messages).Error
		}
	} else {
		// Single message query
		msgNum, _ := strconv.Atoi(sequence)
		offset := msgNum - 1

		err = db.Where("mailbox_id = ?", client.SelectedMboxID).
			Order("id").
			Limit(1).
			Offset(offset).
			Find(&messages).Error
	}

	if err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to fetch messages"})
		return
	}

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Messages fetched", Data: messages})
}
