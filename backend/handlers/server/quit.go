package handlers

import (
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/types"

	"gorm.io/gorm"
)

func handleQuit(db *gorm.DB, client *types.RemoteClient, cmd *models.Command, sendResponse types.SendResponseFunc) {
	client.State = types.StateRemoteDisconnected
	client.Domain = ""
	client.MailFrom = ""
	client.RcptTo = nil

	sendResponse(client.Conn, models.Response{
		ID:      cmd.ID,
		Status:  enums.StatusOK,
		Message: "Goodbye",
	})
}
