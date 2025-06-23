package handlers

import (
	"fmt"
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/repository"
	"pipec-backend/types"
	"time"

	"gorm.io/gorm"
)

type SelectParams struct {
	folder_name string
}

func parseSelectParams(cmd *models.Command, client *types.Client, send types.SendResponseFunc) (*SelectParams, bool) {
	params := &SelectParams{}

	folder_name, ok := cmd.Params["folder"].(string)
	if !ok {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusBAD, Message: "Folder name required"})
		return nil, false
	}

	params.folder_name = folder_name
	return params, true
}

func handleSelect(db *gorm.DB, client *types.Client, cmd *models.Command, send types.SendResponseFunc) {
	if !isAuthenticated(client) {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Not authenticated"})
		return
	}

	params, ok := parseSelectParams(cmd, client, send)
	if !ok {
		return
	}

	fmt.Printf("\033[36mUser %d selecting folder: %s\033[0m\n", client.User.ID, params.folder_name)

	folder, err := repository.GetOrCreateFolderByName(db, client.User.ID, params.folder_name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Folder not found"})
		} else {
			send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Database error"})
		}
		return
	}

	messageCount, err := repository.CountMessagesInFolder(db, folder.ID)
	if err != nil {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Failed to count messages"})
		return
	}

	recentCount, err := repository.CountRecentMessagesInFolder(db, folder.ID, 24*time.Hour)
	if err != nil {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Failed to count recent messages"})
		return
	}

	unseenCount, err := repository.CountUnseenMessagesInFolder(db, folder.ID)
	if err != nil {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Failed to count unseen messages"})
		return
	}

	client.SelectedFolder = folder

	info := models.FolderInfo{
		Name:     params.folder_name,
		Messages: int(messageCount),
		Recent:   int(recentCount),
		Unseen:   int(unseenCount),
	}

	send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusOK, Message: "Folder selected", Data: info})
}
