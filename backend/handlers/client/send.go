package handlers

import (
	"fmt"
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/services"
	"pipec-backend/types"
	"pipec-backend/utils"

	"gorm.io/gorm"
)

// parseSendParams extracts and validates send parameters from command params
func parseSendParams(cmd *models.Command, client *types.Client, sendResponse types.SendResponseFunc) (*types.SendParams, bool) {
	sendParams := &types.SendParams{}

	// Extract "to" field
	to, ok := cmd.Params["to"].([]interface{})
	if !ok || len(to) == 0 {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Missing or invalid 'to' field",
		})
		return nil, false
	}

	sendParams.To = utils.InterfaceToString(to)

	// Extract "cc" field
	cc, ok := cmd.Params["cc"].([]interface{})
	if ok {
		sendParams.CC = utils.InterfaceToString(cc)
	}

	// Extract "bcc" field
	bcc, ok := cmd.Params["bcc"].([]interface{})
	if ok {
		sendParams.BCC = utils.InterfaceToString(bcc)
	}

	// Extract "subject" field
	subject, ok := cmd.Params["subject"].(string)
	if !ok || subject == "" {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Missing or invalid 'subject' field",
		})
		return nil, false
	}

	sendParams.Subject = subject
	// Extract "body" field
	body, ok := cmd.Params["body"].(string)
	if !ok || body == "" {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Missing or invalid 'body' field",
		})
		return nil, false
	}

	sendParams.Body = body

	return sendParams, true
}

// handleSend processes the SEND command for sending emails to multiple recipients
func handleSend(db *gorm.DB, client *types.Client, cmd *models.Command, sendResponse types.SendResponseFunc, localDomain string) {
	// Check if user is authenticated
	if !isAuthenticated(client) {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusNO,
			Message: "Not authenticated",
		})
		return
	}

	// Parse send parameters
	params, ok := parseSendParams(cmd, client, sendResponse)
	if !ok {
		return
	}

	fmt.Printf("\033[32mSending email for user:\033[0m %d \033[34mto:\033[0m %v \033[33mcc:\033[0m %v \033[35mbcc:\033[0m %v\n", client.User.ID, params.To, params.CC, params.BCC)

	address := append(params.To, params.CC...)
	address = append(address, params.BCC...)

	if len(address) == 0 {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "No recipients specified",
		})
		return
	}

	pipeAddress, err := utils.BatchParseEmails(address)
	if err != nil {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: fmt.Sprintf("Invalid email address: %s", err.Error()),
		})
		return
	}

	groupedAddress := utils.GroupEmailsByDomain(pipeAddress)

	var allPassedEmails []string
	var allFailedEmails []string

	for domain, emails := range groupedAddress {
		var passedEmails, failedEmails []string

		if domain == localDomain {
			passedEmails, failedEmails = services.SendMailToLocalServer(db, client, cmd, sendResponse, emails, params)
		} else {
			passedEmails, failedEmails = services.SendMailToRemoteServer(db, client, cmd, sendResponse, domain, emails, params, localDomain)
		}

		allPassedEmails = append(allPassedEmails, passedEmails...)
		allFailedEmails = append(allFailedEmails, failedEmails...)
	}

	// Send response based on results
	if len(allFailedEmails) == 0 {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusOK,
			Message: "All emails sent successfully",
			Data: map[string]interface{}{
				"sent":   len(allPassedEmails),
				"failed": 0,
			},
		})
	} else if len(allPassedEmails) == 0 {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusBAD,
			Message: "Failed to send all emails",
			Data: map[string]interface{}{
				"sent":             0,
				"failed":           len(allFailedEmails),
				"failed_addresses": allFailedEmails,
			},
		})
	} else {
		sendResponse(client.Conn, models.Response{
			ID:      cmd.ID,
			Status:  enums.StatusOK,
			Message: "Partially sent",
			Data: map[string]interface{}{
				"sent":             len(allPassedEmails),
				"failed":           len(allFailedEmails),
				"failed_addresses": allFailedEmails,
			},
		})
	}
}
