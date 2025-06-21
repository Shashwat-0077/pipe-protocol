package server

import (
	"net"
	"pipe-protocol/pkg/protocol"
)

// Handler defines the interface for handling PIPE protocol messages
type Handler interface {
	OnConnect(conn net.Conn) error
	OnHandshake(conn net.Conn, handshake *protocol.Handshake) (*protocol.Response, error)
	OnSend(conn net.Conn, send *protocol.Send) (*protocol.Response, error)
	OnDisconnect(conn net.Conn) error
	OnError(conn net.Conn, err error)
}
