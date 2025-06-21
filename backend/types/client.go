package types

import "net"

type ConnectionState int

const (
	StateNotAuthenticated ConnectionState = iota
	StateAuthenticated
	StateSelected
	StateLogout
)

type Client struct {
	Conn           net.Conn
	State          ConnectionState
	UserID         int
	Username       string
	SelectedMbox   string
	SelectedMboxID int
}
