package handlers

import (
	"fmt"
	"pipec-backend/models"
	"pipec-backend/types"

	"gorm.io/gorm"
)

func handleLogout(_ *gorm.DB, client *types.Client, cmd *models.Command, send types.SendResponseFunc) {
	client.State = types.StateLogout

	fmt.Printf("\033[31mClient logged out: %d\033[0m\n", client.User.ID)

	send(client.Conn, models.Response{ID: cmd.ID, Status: "OK", Message: "Logout successful"})
}
