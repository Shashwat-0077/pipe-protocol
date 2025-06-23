package handlers

import (
	"pipec-backend/models"
	"pipec-backend/types"
	"strings"

	"gorm.io/gorm"
)

func HandleCommand(db *gorm.DB, rc *types.RemoteClient, cmd *models.Command, send types.SendResponseFunc, localDomain string) {

	switch strings.ToUpper(cmd.Command) {
	case "HELO":
		handleHelo(db, rc, cmd, send)
	case "MAIL_FROM":
		handleMailFrom(db, rc, cmd, send)
	case "RCPT_TO":
		handleRcptTo(db, rc, cmd, send)
	case "DATA":
		handleData(db, rc, cmd, send)
	case "QUIT":
		handleQuit(db, rc, cmd, send)

	default:
		send(rc.Conn, models.Response{ID: cmd.ID, Status: "BAD", Message: "Unknown command for server"})
	}
}
