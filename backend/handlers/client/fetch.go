package handlers

import (
	"fmt"
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/repository"
	"pipec-backend/types"
	"pipec-backend/utils"

	"gorm.io/gorm"
)

type FetchParams struct {
	sequence string
}

func parseFetchParams(cmd *models.Command, client *types.Client, send types.SendResponseFunc) (*FetchParams, bool) {
	params := &FetchParams{}

	// Extract "sequence" field
	seq, ok := cmd.Params["sequence"].(string)
	if !ok || seq == "" {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusBAD, Message: "Sequence number required"})
		return nil, false
	}

	params.sequence = seq

	return params, true
}

func handleFetch(db *gorm.DB, client *types.Client, cmd *models.Command, send types.SendResponseFunc) {
	if !isAuthenticated(client) {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "User is not authenticated"})
		return
	}

	if client.SelectedFolder == nil {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "No folder selected"})
		return
	}

	params, ok := parseFetchParams(cmd, client, send)
	if !ok {
		return
	}

	fmt.Printf("\033[32mFetching messages for user: %d in folder: %s sequence: %s\033[0m\n", client.User.ID, client.SelectedFolder.Name, params.sequence)

	start, end, err := utils.ParseSequence(params.sequence)
	if err != nil {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusBAD, Message: "Invalid sequence format"})
		return
	}

	if start == -1 && end == -1 {
		// "*" — fetch latest message
		var msg models.Message
		err := db.Where("folder_id = ?", client.SelectedFolder).
			Order("id DESC").Limit(1).First(&msg).Error

		if err != nil {
			send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "No messages found"})
			return
		}
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusOK, Message: "Fetched latest message", Data: []models.Message{msg}})
		return
	}

	if start == end {
		msg, err := repository.GetMessageBySequence(db, client.SelectedFolder.ID, start)
		if err != nil {
			send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Message not found"})
			return
		}
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusOK, Message: "Message fetched", Data: []models.Message{*msg}})
		return
	}

	msgs, err := repository.GetMessagesBySequenceRange(db, client.SelectedFolder.ID, start, end)
	if err != nil {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Failed to fetch messages"})
		return
	}
	send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusOK, Message: "Messages fetched", Data: msgs})
}
