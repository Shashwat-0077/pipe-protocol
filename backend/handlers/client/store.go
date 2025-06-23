package handlers

import (
	"fmt"
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/repository"
	"pipec-backend/types"
	"strconv"

	"gorm.io/gorm"
)

type StoreParams struct {
	sequence string
	flags    []string
}

func parseStoreParams(cmd *models.Command, client *types.Client, send types.SendResponseFunc) (*StoreParams, bool) {
	params := &StoreParams{}

	sequence, ok := cmd.Params["sequence"].(string)
	if !ok {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusBAD, Message: "Sequence number required"})
		return nil, false
	}

	flagsInterface, ok := cmd.Params["flags"].([]interface{})
	if !ok {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusBAD, Message: "Flags must be an array"})
		return nil, false
	}

	var flags []string
	for _, flag := range flagsInterface {
		if str, ok := flag.(string); ok {
			flags = append(flags, str)
		}
	}

	params.sequence = sequence
	params.flags = flags

	return params, true
}

func handleStore(db *gorm.DB, client *types.Client, cmd *models.Command, send types.SendResponseFunc) {
	if !isAuthenticated(client) {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "No folder selected"})
		return
	}

	params, ok := parseStoreParams(cmd, client, send)
	if !ok {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusBAD, Message: "Invalid parameters"})
		return
	}

	fmt.Printf("\033[32mStoring flags for user: %d in folder: %s sequence: %s flags: %v\033[0m\n", client.User.ID, client.SelectedFolder.Name, params.sequence, params.flags)

	msgNum, err := strconv.Atoi(params.sequence)
	if err != nil || msgNum <= 0 {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusBAD, Message: "Invalid sequence number"})
		return
	}

	message, err := repository.GetMessageBySequence(db, uint(client.SelectedFolder.ID), msgNum)
	if err != nil {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Message not found"})
		return
	}

	err = repository.UpdateMessageFlags(db, message.ID, params.flags)
	if err != nil {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Failed to store flags"})
		return
	}

	send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusOK, Message: "Flags stored"})
}
