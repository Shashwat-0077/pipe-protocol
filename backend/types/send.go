package types

import (
	"net"
	"pipec-backend/models"
)

type SendResponseFunc func(client net.Conn, resp models.Response)

type SendParams struct {
	To      []string `json:"to"`
	CC      []string `json:"cc,omitempty"`
	BCC     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}
