package handlers

import (
	"pipec-backend/models"
	"pipec-backend/types"
	"strings"

	"gorm.io/gorm"
)

func HandleCommand(db *gorm.DB, client *types.Client, cmd *models.Command, send types.SendResponseFunc, localDomain string) {

	switch strings.ToUpper(cmd.Command) {
	case "LOGIN":
		handleLogin(db, client, cmd, send)
	case "LIST":
		handleList(db, client, cmd, send)
	case "SELECT":
		handleSelect(db, client, cmd, send)
	case "FETCH":
		handleFetch(db, client, cmd, send)
	case "SEND":
		handleSend(db, client, cmd, send, localDomain)
	case "STORE":
		handleStore(db, client, cmd, send)
	case "LOGOUT":
		handleLogout(db, client, cmd, send)

	default:
		send(client.Conn, models.Response{ID: cmd.ID, Status: "BAD", Message: "Unknown command for client"})
	}
}

func isAuthenticated(client *types.Client) bool {
	return client.State == types.StateAuthenticated
}
