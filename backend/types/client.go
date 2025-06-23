package types

import (
	"net"
	"pipec-backend/enums"
	"pipec-backend/models"
)

type BaseConnection struct {
	Conn           net.Conn             `json:"conn"`
	ConnectionType enums.ConnectionType `json:"connection_type"`
}

type ClientState int

const (
	StateNotAuthenticated ClientState = iota
	StateAuthenticated
	StateLogout
)

type Client struct {
	BaseConnection
	State          ClientState    `json:"state"`
	User           *models.User   `json:"user,omitempty"`
	SelectedFolder *models.Folder `json:"selected_folder,omitempty"`
}

type RemoteClientState int

const (
	StateRemoteDisconnected RemoteClientState = iota
	StateRemoteConnected
	StateRemoteMailFrom
	StateRemoteRcptTo
	StateRemoteData
	StateRemoteQuit
)

type RemoteClient struct {
	BaseConnection
	State    RemoteClientState `json:"state"`
	MailFrom string            `json:"mail_from,omitempty"`
	RcptTo   []string          `json:"rcpt_to,omitempty"`
	Domain   string            `json:"domain,omitempty"`
}
