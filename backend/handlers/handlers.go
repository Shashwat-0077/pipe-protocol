package handlers

import (
	"pipec-backend/models"
	"pipec-backend/types"
	"strings"

	"gorm.io/gorm"
)

type SendResponseFunc func(client *types.Client, resp models.Response)

func HandleCommand(db *gorm.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
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
		handleSend(db, client, cmd, send)
	case "STORE":
		handleStore(db, client, cmd, send)
	case "LOGOUT":
		handleLogout(db, client, cmd, send)
	default:
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Unknown command"})
	}
}
