package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"pipec-backend/enums"
	clientHandlers "pipec-backend/handlers/client"
	serverHandlers "pipec-backend/handlers/server"
	"pipec-backend/models"
	"pipec-backend/types"
	"pipec-backend/utils"

	"gorm.io/gorm"
)

type User = models.User

type PipecServer struct {
	db          *gorm.DB // Changed from *sql.DB to *gorm.DB
	listener    net.Listener
	localDomain string
	port        string
}

func NewPipeCServer(db *gorm.DB, port string) *PipecServer { // Changed parameter type
	return &PipecServer{
		db:          db,
		localDomain: "pipec.local",
		port:        port,
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

func (s *PipecServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("New connection from: %s", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)

	var connType enums.ConnectionType
	var client *types.Client
	var remote *types.RemoteClient
	var connectionTypeResolved bool

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var cmd models.Command
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			// Temporary connection type fallback for bad JSON
			utils.SendResponse(conn, models.Response{ID: "", Status: "BAD", Message: "Invalid JSON format"})
			continue
		}

		if !connectionTypeResolved {
			connType = cmd.ConnectionType

			if !connType.IsValid() {
				utils.SendResponse(conn, models.Response{ID: cmd.ID, Status: "BAD", Message: "Unknown connection type"})
				return
			}

			connectionTypeResolved = true

			switch connType {
			case enums.ConnectionTypeClient:
				client = &types.Client{
					BaseConnection: types.BaseConnection{
						Conn:           conn,
						ConnectionType: connType,
					},
					State: types.StateNotAuthenticated,
				}
			case enums.ConnectionTypeServer:
				remote = &types.RemoteClient{
					BaseConnection: types.BaseConnection{
						Conn:           conn,
						ConnectionType: connType,
					},
					State: types.StateRemoteConnected,
				}
			}
		}

		switch connType {
		case enums.ConnectionTypeClient:
			clientHandlers.HandleCommand(s.db, client, &cmd, utils.SendResponse, s.localDomain)
			if client.State == types.StateLogout {
				log.Printf("Client logged out: %s", conn.RemoteAddr())
				return
			}
		case enums.ConnectionTypeServer:
			serverHandlers.HandleCommand(s.db, remote, &cmd, utils.SendResponse, s.localDomain)
			if remote.State == types.StateRemoteQuit {
				log.Printf("Remote server quit: %s", conn.RemoteAddr())
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Connection error from %s: %v", conn.RemoteAddr(), err)
	}

	log.Printf("Connection closed: %s", conn.RemoteAddr())

	// TODO: Implement inactivity-based auto-close (use time.AfterFunc / timers)
	// BUG: Still subject to race conditions if connection reuse or shared state isn't guarded by mutex
}
