package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/repository"
	"pipec-backend/types"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// SendMailToLocalServer handles sending emails to users on the local domain
func SendMailToLocalServer(db *gorm.DB, client *types.Client, cmd *models.Command,
	sendResponse types.SendResponseFunc, emails []models.PipeAddress, params *types.SendParams) ([]string, []string) {

	var passedEmails []string
	var failedEmails []string

	for _, email := range emails {
		// Find the user by username (email.Username)
		user, err := repository.GetUserByUsername(db, email.Username)
		if err != nil {
			failedEmails = append(failedEmails, email.Raw)
			continue
		}

		// Get or create INBOX folder for the user
		folder, err := repository.GetOrCreateFolderByName(db, user.ID, "INBOX")
		if err != nil {
			failedEmails = append(failedEmails, email.Raw)
			continue
		}

		// Create the message
		message := &models.Message{
			UserID:   user.ID,
			FolderID: folder.ID,
			From:     client.User.Email, // Assuming client has these fields
			To:       pq.StringArray(params.To),
			Cc:       pq.StringArray(params.CC),
			Bcc:      pq.StringArray(params.BCC),
			Subject:  params.Subject,
			Body:     params.Body,
			Flags:    pq.StringArray{},        // New message, no flags initially
			UID:      uint(time.Now().Unix()), // Simple UID generation
		}

		// Save the message to database
		err = repository.CreateMessage(db, message)
		if err != nil {
			failedEmails = append(failedEmails, email.Raw)
			continue
		}

		passedEmails = append(passedEmails, email.Raw)
	}

	return passedEmails, failedEmails
}

// SendMailToRemoteServer handles sending emails to remote PIPE servers
func SendMailToRemoteServer(db *gorm.DB, client *types.Client, cmd *models.Command,
	sendResponse types.SendResponseFunc, domain string, emails []models.PipeAddress, params *types.SendParams, localDomain string) ([]string, []string) {

	var passedEmails []string
	var failedEmails []string

	// Get the best PIPE server for the domain
	connectResult, err := QuickConnect(domain)
	if err != nil {
		// If connection fails, all emails for this domain fail
		for _, email := range emails {
			failedEmails = append(failedEmails, email.Raw)
		}
		return passedEmails, failedEmails
	}

	// Establish connection to remote server
	conn, err := net.Dial("tcp", net.JoinHostPort(connectResult.Server.Host, fmt.Sprintf("%d", connectResult.Server.Port)))
	if err != nil {
		for _, email := range emails {
			failedEmails = append(failedEmails, email.Raw)
		}
		return passedEmails, failedEmails
	}
	defer conn.Close()

	// Perform PIPE handshake and send emails
	success := performRemoteHandshake(conn, localDomain, emails, params, client)

	if success {
		for _, email := range emails {
			passedEmails = append(passedEmails, email.Raw)
		}
	} else {
		for _, email := range emails {
			failedEmails = append(failedEmails, email.Raw)
		}
	}

	return passedEmails, failedEmails
}

// performRemoteHandshake handles the PIPE protocol communication with remote server
func performRemoteHandshake(conn net.Conn, localDomain string, emails []models.PipeAddress, params *types.SendParams, client *types.Client) bool {
	// Create a buffered reader for reading responses
	reader := bufio.NewReader(conn)

	// Generate unique command IDs
	cmdID := 1

	// Send HELO command and wait for response
	heloCmd := models.Command{
		ID:             fmt.Sprintf("cmd_%d", cmdID),
		Command:        "HELO",
		ConnectionType: enums.ConnectionTypeServer,
		Params: map[string]interface{}{
			"domain": localDomain,
		},
	}

	if !sendCommandAndWaitForAck(conn, reader, heloCmd) {
		return false
	}
	cmdID++

	// Send MAIL_FROM command and wait for response
	mailFromCmd := models.Command{
		ID:             fmt.Sprintf("cmd_%d", cmdID),
		Command:        "MAIL_FROM",
		ConnectionType: enums.ConnectionTypeServer,
		Params: map[string]interface{}{
			"from": client.User.Email,
		},
	}

	if !sendCommandAndWaitForAck(conn, reader, mailFromCmd) {
		return false
	}
	cmdID++

	// Send RCPT_TO commands for each recipient and wait for responses
	for _, email := range emails {
		rcptCmd := models.Command{
			ID:             fmt.Sprintf("cmd_%d", cmdID),
			Command:        "RCPT_TO",
			ConnectionType: enums.ConnectionTypeServer,
			Params: map[string]interface{}{
				"to": email.Raw,
			},
		}

		if !sendCommandAndWaitForAck(conn, reader, rcptCmd) {
			return false
		}
		cmdID++
	}

	// Send email content with DATA command and wait for response
	dataCmd := models.Command{
		ID:             fmt.Sprintf("cmd_%d", cmdID),
		Command:        "DATA",
		ConnectionType: enums.ConnectionTypeServer,
		Params: map[string]interface{}{
			"subject": params.Subject,
			"body":    params.Body,
			"to":      params.To,
			"cc":      params.CC,
			"bcc":     params.BCC,
		},
	}

	if !sendCommandAndWaitForAck(conn, reader, dataCmd) {
		return false
	}
	cmdID++

	// Send QUIT command and wait for response
	quitCmd := models.Command{
		ID:             fmt.Sprintf("cmd_%d", cmdID),
		Command:        "QUIT",
		ConnectionType: enums.ConnectionTypeServer,
		Params:         map[string]interface{}{},
	}

	if !sendCommandAndWaitForAck(conn, reader, quitCmd) {
		return false
	}

	return true
}

// sendCommandAndWaitForAck sends a command and waits for the acknowledgment
func sendCommandAndWaitForAck(conn net.Conn, reader *bufio.Reader, cmd models.Command) bool {
	// Send the command
	if !sendJSONCommand(conn, cmd) {
		return false
	}

	// Wait for response
	response, err := readResponse(conn, reader)
	if err != nil {
		return false
	}

	// Check if the response ID matches the command ID
	if response.ID != cmd.ID {
		return false
	}

	// Check if the response status indicates success
	return response.Status == enums.StatusOK
}

// readResponse reads a JSON response from the connection
func readResponse(conn net.Conn, reader *bufio.Reader) (*models.Response, error) {
	// Set a read timeout to prevent hangs
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Read line from connection
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var response models.Response
	err = json.Unmarshal(line, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// sendJSONCommand sends a JSON command to the connection
func sendJSONCommand(conn net.Conn, cmd models.Command) bool {
	// Encode the command into JSON
	data, err := json.Marshal(cmd)
	if err != nil {
		return false
	}

	// Append newline to indicate message boundary
	data = append(data, '\n')

	// Set a write timeout to prevent hangs
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	// Send the command
	_, err = conn.Write(data)

	return err == nil
}
