package models

import "pipec-backend/enums"

type Command struct {
	ID             string                 `json:"id"`
	Command        string                 `json:"command"`
	ConnectionType enums.ConnectionType   `json:"connection_type"`
	Params         map[string]interface{} `json:"params"`
}

type Response struct {
	ID      string       `json:"id"`
	Status  enums.Status `json:"status"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
}
