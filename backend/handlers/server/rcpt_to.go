package handlers

import (
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/repository"
	"pipec-backend/types"
	"pipec-backend/utils"

	"gorm.io/gorm"
)

func handleRcptTo(db *gorm.DB, client *types.RemoteClient, cmd *models.Command, sendResponse types.SendResponseFunc) {
	if client.State != types.StateRemoteMailFrom && client.State != types.StateRemoteRcptTo {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Must send MAIL_FROM first",
		})
		return
	}

	to, ok := cmd.Params["to"].(string)
	if !ok || to == "" {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Missing recipient address",
		})
		return
	}

	// Parse and validate the recipient address
	pipeAddr, err := utils.ParseEmail(to)
	if err != nil {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Invalid recipient address",
		})
		return
	}

	// Check if recipient exists locally
	_, err = repository.GetUserByUsername(db, pipeAddr.Username)
	if err != nil {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Recipient not found",
		})
		return
	}

	// Add to recipients list
	if client.RcptTo == nil {
		client.RcptTo = make([]string, 0)
	}
	client.RcptTo = append(client.RcptTo, to)
	client.State = types.StateRemoteRcptTo

	sendResponse(client.Conn, models.Response{
		ID:      cmd.ID,
		Status:  enums.StatusOK,
		Message: "Recipient accepted",
	})
}
