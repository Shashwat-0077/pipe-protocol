package server

import (
	"fmt"
	"log"
	"net"
	"pipe-protocol/pkg/protocol"
	"pipe-protocol/pkg/storage"
	"strings"
)

// PipeHandler implements the Handler interface
type PipeHandler struct {
	serverName string
	storage    *storage.FileStorage
}

// NewPipeHandler creates a new PIPE handler
func NewPipeHandler(serverName string, storage *storage.FileStorage) *PipeHandler {
	return &PipeHandler{
		serverName: serverName,
		storage:    storage,
	}
}

// OnConnect handles new connections
func (h *PipeHandler) OnConnect(conn net.Conn) error {
	log.Printf("New connection from %s", conn.RemoteAddr())
	return nil
}

// OnHandshake handles handshake messages
func (h *PipeHandler) OnHandshake(conn net.Conn, handshake *protocol.Handshake) (*protocol.Response, error) {
	log.Printf("Handshake from %s (version: %s)", handshake.ServerName, handshake.Version)

	// Validate handshake
	if handshake.Version != "1.0" {
		return protocol.NewResponse(
			handshake.ID,
			protocol.StatusError,
			"Unsupported protocol version",
			nil,
		), nil
	}

	if handshake.ConnectionType != "server" {
		return protocol.NewResponse(
			handshake.ID,
			protocol.StatusError,
			"Invalid connection type",
			nil,
		), nil
	}

	return protocol.NewResponse(
		handshake.ID,
		protocol.StatusOK,
		fmt.Sprintf("Handshake accepted by %s", h.serverName),
		nil,
	), nil
}

// OnSend handles send messages
func (h *PipeHandler) OnSend(conn net.Conn, send *protocol.Send) (*protocol.Response, error) {
	log.Printf("Received message from %s to %v: %s", send.From, send.To, send.Subject)

	var accepted []string
	var rejected []string
	var reason string

	// Validate recipients
	for _, recipient := range send.To {
		if h.isLocalRecipient(recipient) {
			// Store the message locally
			msg := &storage.EmailMessage{
				ID:      send.ID,
				From:    send.From,
				To:      []string{recipient},
				Subject: send.Subject,
				Body:    send.Body,
				Headers: send.Headers,
			}

			if err := h.storage.StoreIncoming(msg); err != nil {
				log.Printf("Failed to store message for %s: %v", recipient, err)
				rejected = append(rejected, recipient)
				reason = "Storage error"
			} else {
				accepted = append(accepted, recipient)
			}
		} else {
			rejected = append(rejected, recipient)
			reason = "Not a local recipient"
		}
	}

	// Determine response status
	status := protocol.StatusOK
	message := "Message accepted for delivery"

	if len(accepted) == 0 {
		status = protocol.StatusError
		message = "No valid recipients"
	} else if len(rejected) > 0 {
		message = "Message partially accepted"
	}

	details := &protocol.ResponseDetails{
		Accepted: accepted,
		Rejected: rejected,
		Reason:   reason,
	}

	return protocol.NewResponse(send.ID, status, message, details), nil
}

// OnDisconnect handles connection disconnection
func (h *PipeHandler) OnDisconnect(conn net.Conn) error {
	log.Printf("Connection closed from %s", conn.RemoteAddr())
	return nil
}

// OnError handles errors
func (h *PipeHandler) OnError(conn net.Conn, err error) {
	log.Printf("Error from %s: %v", conn.RemoteAddr(), err)
}

// isLocalRecipient checks if a recipient is local to this server
func (h *PipeHandler) isLocalRecipient(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]
	// In a real implementation, you'd check against configured local domains
	// For now, we'll check if the domain matches our server name
	return strings.Contains(h.serverName, domain) || domain == "localhost" || domain == "local"
}
