// 📁 handlers/list.go
package handlers

import (
	"fmt"
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/repository"
	"pipec-backend/types"

	"gorm.io/gorm"
)

func handleList(db *gorm.DB, client *types.Client, cmd *models.Command, send types.SendResponseFunc) {
	if !isAuthenticated(client) {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Not authenticated"})
		return
	}

	fmt.Printf("\033[32mListing folders for user: %d\033[0m\n", client.User.ID)

	folders, err := repository.GetFoldersByUser(db, client.User.ID)
	if err != nil {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Failed to list folders"})
		return
	}

	send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusOK, Message: "Folders listed", Data: folders})
}
