package handlers

import (
	"pipec-backend/models"
	"pipec-backend/types"

	"gorm.io/gorm"
)

func handleLogout(_ *gorm.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	client.State = types.StateLogout
	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Logout successful"})
}
