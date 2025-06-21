package handlers

import (
	"pipec-backend/models"
	"pipec-backend/types"
	"strconv"

	"gorm.io/gorm"
)

func handleStore(db *gorm.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State != types.StateSelected {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "No mailbox selected"})
		return
	}

	sequence, ok := cmd.Params["sequence"].(string)
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Sequence required"})
		return
	}

	flagsInterface, ok := cmd.Params["flags"].([]interface{})
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Flags required"})
		return
	}

	var flags []string
	for _, flag := range flagsInterface {
		if str, ok := flag.(string); ok {
			flags = append(flags, str)
		}
	}

	msgNum, _ := strconv.Atoi(sequence)

	// Find the message at the specified sequence number
	var message models.Message
	err := db.Where("mailbox_id = ?", client.SelectedMboxID).
		Order("id").
		Offset(msgNum - 1).
		Limit(1).
		First(&message).Error

	if err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Message not found"})
		return
	}

	// Update the flags
	err = db.Model(&message).Update("flags", models.JSON(flags)).Error
	if err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to store flags"})
		return
	}

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Flags stored"})
}
