package handlers

import (
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/types"

	"gorm.io/gorm"
)

func handleMailFrom(db *gorm.DB, client *types.RemoteClient, cmd *models.Command, sendResponse types.SendResponseFunc) {
	if client.State != types.StateRemoteConnected {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Must send HELO first",
		})
		return
	}

	from, ok := cmd.Params["from"].(string)
	if !ok || from == "" {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Missing from address",
		})
		return
	}

	// Store the sender address
	client.MailFrom = from
	client.State = types.StateRemoteMailFrom

	sendResponse(client.Conn, models.Response{
		ID:      cmd.ID,
		Status:  enums.StatusOK,
		Message: "Mail from accepted",
	})
}
