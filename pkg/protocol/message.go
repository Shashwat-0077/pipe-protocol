package protocol

import (
	"encoding/json"
	"time"
)

// MessageType defines the type of protocol message
type MessageType string

const (
	TypeHandshake MessageType = "handshake"
	TypeSend      MessageType = "send"
	TypeResponse  MessageType = "response"
)

// Status defines response status
type Status string

const (
	StatusOK    Status = "OK"
	StatusError Status = "ERROR"
	StatusRetry Status = "RETRY"
)

// BaseMessage contains common fields for all messages
type BaseMessage struct {
	Type      MessageType `json:"type"`
	ID        string      `json:"id"`
	Timestamp int64       `json:"timestamp"`
}

// Handshake message for initial connection
type Handshake struct {
	BaseMessage
	ConnectionType string `json:"connection_type"`
	Version        string `json:"version"`
	ServerName     string `json:"server_name"`
}

// Send message for email transmission
type Send struct {
	BaseMessage
	From    string            `json:"from"`
	To      []string          `json:"to"`
	Subject string            `json:"subject"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

// ResponseDetails contains detailed response information
type ResponseDetails struct {
	Accepted []string `json:"accepted,omitempty"`
	Rejected []string `json:"rejected,omitempty"`
	Reason   string   `json:"reason,omitempty"`
}

// Response message for acknowledgments
type Response struct {
	BaseMessage
	ReplyTo string           `json:"reply_to"`
	Status  Status           `json:"status"`
	Message string           `json:"message"`
	Details *ResponseDetails `json:"details,omitempty"`
}

// ParseMessage parses raw JSON into appropriate message type
func ParseMessage(data []byte) (interface{}, error) {
	var base BaseMessage
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.Type {
	case TypeHandshake:
		var msg Handshake
		err := json.Unmarshal(data, &msg)
		return &msg, err
	case TypeSend:
		var msg Send
		err := json.Unmarshal(data, &msg)
		return &msg, err
	case TypeResponse:
		var msg Response
		err := json.Unmarshal(data, &msg)
		return &msg, err
	default:
		return nil, json.Unmarshal(data, &base)
	}
}

// ToJSON converts message to JSON bytes
func ToJSON(msg interface{}) ([]byte, error) {
	return json.Marshal(msg)
}

// NewHandshake creates a new handshake message
func NewHandshake(serverName string) *Handshake {
	return &Handshake{
		BaseMessage: BaseMessage{
			Type:      TypeHandshake,
			ID:        generateID(),
			Timestamp: time.Now().Unix(),
		},
		ConnectionType: "server",
		Version:        "1.0",
		ServerName:     serverName,
	}
}

// NewSend creates a new send message
func NewSend(from string, to []string, subject, body string, headers map[string]string) *Send {
	if headers == nil {
		headers = make(map[string]string)
	}

	return &Send{
		BaseMessage: BaseMessage{
			Type:      TypeSend,
			ID:        generateID(),
			Timestamp: time.Now().Unix(),
		},
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
		Headers: headers,
	}
}

// NewResponse creates a new response message
func NewResponse(replyTo string, status Status, message string, details *ResponseDetails) *Response {
	return &Response{
		BaseMessage: BaseMessage{
			Type:      TypeResponse,
			ID:        generateID(),
			Timestamp: time.Now().Unix(),
		},
		ReplyTo: replyTo,
		Status:  status,
		Message: message,
		Details: details,
	}
}

// generateID generates a unique message ID
func generateID() string {
	// Simple ID generation - in production, use UUID or similar
	return string(rune(time.Now().UnixNano()))
}
