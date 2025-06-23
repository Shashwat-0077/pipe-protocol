package handlers

import (
	"fmt"
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/types"

	"gorm.io/gorm"
)

func handleHelo(db *gorm.DB, client *types.RemoteClient, cmd *models.Command, sendResponse types.SendResponseFunc) {
	domain, ok := cmd.Params["domain"].(string)
	if !ok || domain == "" {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Missing domain in HELO command",
		})
		return
	}

	// Store the remote domain for this connection
	client.Domain = domain
	client.State = types.StateRemoteConnected

	sendResponse(client.Conn, models.Response{
		ID:      cmd.ID,
		Status:  enums.StatusOK,
		Message: fmt.Sprintf("Hello %s", domain),
	})
}
