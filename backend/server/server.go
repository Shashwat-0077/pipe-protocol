package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"pipec-backend/handlers"
	"pipec-backend/models"
	"pipec-backend/types"

	"gorm.io/gorm"
)

type User = models.User

type PipecServer struct {
	db       *gorm.DB // Changed from *sql.DB to *gorm.DB
	listener net.Listener
	port     string
}

func NewPipeCServer(db *gorm.DB, port string) *PipecServer { // Changed parameter type
	return &PipecServer{
		db:   db,
		port: port,
	}
}

func (s *PipecServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", s.port, err)
	}
	log.Printf("PIPEC server listening on port %s", s.port)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// TODO : Implement auto closing of connections after inactivity
// BUG : Possible race condition if multiple clients connect at the same time
func (s *PipecServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	client := &types.Client{Conn: conn, State: types.StateNotAuthenticated}
	log.Printf("New client connected: %s", conn.RemoteAddr())
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var cmd models.Command
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			s.sendResponse(client, models.Response{ID: "", Status: "BAD", Message: "Invalid JSON format"})
			continue
		}
		handlers.HandleCommand(s.db, client, &cmd, s.sendResponse) // Pass *gorm.DB
		if client.State == types.StateLogout {
			break
		}
	}
	log.Printf("Client disconnected: %s", conn.RemoteAddr())
}

func (s *PipecServer) sendResponse(client *types.Client, resp models.Response) {
	data, _ := json.Marshal(resp)
	client.Conn.Write(append(data, '\n'))
}
