package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"pipe-protocol/pkg/protocol"
	"strings"
	"time"
)

// Server represents the PIPE server
type Server struct {
	host       string
	port       int
	handler    Handler
	timeout    time.Duration
	listener   net.Listener
	maxConn    int
	activeConn int
}

// NewServer creates a new PIPE server
func NewServer(host string, port int, handler Handler, timeout time.Duration, maxConn int) *Server {
	return &Server{
		host:    host,
		port:    port,
		handler: handler,
		timeout: timeout,
		maxConn: maxConn,
	}
}

// Start starts the server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	log.Printf("PIPE server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		if s.activeConn >= s.maxConn {
			log.Printf("Max connections reached, rejecting %s", conn.RemoteAddr())
			conn.Close()
			continue
		}

		s.activeConn++
		go s.handleConnection(conn)
	}
}

// Stop stops the server
func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handleConnection handles a single connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.activeConn--
		conn.Close()
		s.handler.OnDisconnect(conn)
	}()

	// Set connection timeout
	conn.SetDeadline(time.Now().Add(s.timeout))

	if err := s.handler.OnConnect(conn); err != nil {
		s.handler.OnError(conn, err)
		return
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Read messages until connection closes or error
	for {
		// Read line (JSON message)
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				s.handler.OnError(conn, err)
			}
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse message
		msg, err := protocol.ParseMessage([]byte(line))
		if err != nil {
			s.handler.OnError(conn, fmt.Errorf("failed to parse message: %w", err))
			s.sendError(writer, "", "Invalid message format")
			continue
		}

		// Handle message based on type
		var response *protocol.Response

		switch m := msg.(type) {
		case *protocol.Handshake:
			response, err = s.handler.OnHandshake(conn, m)
		case *protocol.Send:
			response, err = s.handler.OnSend(conn, m)
		default:
			err = fmt.Errorf("unsupported message type")
		}

		if err != nil {
			s.handler.OnError(conn, err)
			s.sendError(writer, "", err.Error())
			continue
		}

		// Send response
		if response != nil {
			if err := s.sendResponse(writer, response); err != nil {
				s.handler.OnError(conn, err)
				break
			}
		}

		// Reset timeout for next message
		conn.SetDeadline(time.Now().Add(s.timeout))
	}
}

// sendResponse sends a response message
func (s *Server) sendResponse(writer *bufio.Writer, response *protocol.Response) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	_, err = writer.Write(append(data, '\n'))
	if err != nil {
		return err
	}

	return writer.Flush()
}

// sendError sends an error response
func (s *Server) sendError(writer *bufio.Writer, replyTo, message string) {
	response := protocol.NewResponse(replyTo, protocol.StatusError, message, nil)
	s.sendResponse(writer, response)
}
