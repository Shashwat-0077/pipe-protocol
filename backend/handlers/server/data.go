package handlers

import (
	"fmt"
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/repository"
	"pipec-backend/types"
	"pipec-backend/utils"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

func handleData(db *gorm.DB, client *types.RemoteClient, cmd *models.Command, sendResponse types.SendResponseFunc) {
	if client.State != types.StateRemoteRcptTo {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Must send DATA first",
		})
		return
	}

	// Get email data from JSON params
	subject, ok := cmd.Params["subject"].(string)
	if !ok {
		subject = ""
	}

	body, ok := cmd.Params["body"].(string)
	if !ok {
		body = ""
	}

	// Get recipient arrays
	var toRecipients, ccRecipients, bccRecipients []string

	if to, ok := cmd.Params["to"].([]interface{}); ok {
		toRecipients = utils.InterfaceToString(to)
	}

	if cc, ok := cmd.Params["cc"].([]interface{}); ok {
		ccRecipients = utils.InterfaceToString(cc)
	}

	if bcc, ok := cmd.Params["bcc"].([]interface{}); ok {
		bccRecipients = utils.InterfaceToString(bcc)
	}

	// Deliver email to each recipient
	var failedCount int
	var deliveredCount int

	for _, rcpt := range client.RcptTo {
		pipeAddr, err := utils.ParseEmail(rcpt)
		if err != nil {
			failedCount++
			continue
		}

		user, err := repository.GetUserByUsername(db, pipeAddr.Username)
		if err != nil {
			failedCount++
			continue
		}

		folder, err := repository.GetOrCreateFolderByName(db, user.ID, "INBOX")
		if err != nil {
			failedCount++
			continue
		}

		message := &models.Message{
			UserID:   user.ID,
			FolderID: folder.ID,
			From:     client.MailFrom,
			To:       pq.StringArray(toRecipients),
			Cc:       pq.StringArray(ccRecipients),
			Bcc:      pq.StringArray(bccRecipients),
			Subject:  subject,
			Body:     body,
			Flags:    pq.StringArray{},
			UID:      uint(time.Now().Unix()),
		}

		err = repository.CreateMessage(db, message)
		if err != nil {
			failedCount++
		} else {
			deliveredCount++
		}
	}

	// Reset client state for next transaction
	client.State = types.StateRemoteConnected
	client.MailFrom = ""
	client.RcptTo = nil

	if failedCount > 0 && deliveredCount == 0 {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Failed to deliver to all recipients",
		})
	} else if failedCount > 0 {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusOK,
			Message: fmt.Sprintf("Delivered to %d recipients, failed %d", deliveredCount, failedCount),
		})
	} else {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusOK,
			Message: "Message delivered successfully",
		})
	}
}
