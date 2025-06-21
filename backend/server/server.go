package server

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"pipec-backend/models"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Import types from models package
type User = models.User
type Message = models.Message
type Mailbox = models.Mailbox
type MailboxInfo = models.MailboxInfo
type Command = models.Command
type Response = models.Response

// Server
type PipecServer struct {
	db       *sql.DB
	listener net.Listener
	port     string
}

// Connection states
type ConnectionState int

const (
	StateNotAuthenticated ConnectionState = iota
	StateAuthenticated
	StateSelected
	StateLogout
)

// Client connection
type Client struct {
	conn           net.Conn
	state          ConnectionState
	userID         int
	username       string
	selectedMbox   string
	selectedMboxID int
}

func NewPipeCServer(db *sql.DB, port string) *PipecServer {
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

func (s *PipecServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	client := &Client{
		conn:  conn,
		state: StateNotAuthenticated,
	}

	log.Printf("New client connected: %s", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var cmd Command
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			s.sendResponse(client, Response{
				ID:      "",
				Status:  "BAD",
				Message: "Invalid JSON format",
			})
			continue
		}

		s.handleCommand(client, cmd)

		if client.state == StateLogout {
			break
		}
	}

	log.Printf("Client disconnected: %s", conn.RemoteAddr())
}

func (s *PipecServer) handleCommand(client *Client, cmd Command) {
	switch strings.ToUpper(cmd.Command) {
	case "LOGIN":
		s.handleLogin(client, cmd)
	case "LIST":
		s.handleList(client, cmd)
	case "SELECT":
		s.handleSelect(client, cmd)
	case "FETCH":
		s.handleFetch(client, cmd)
	case "SEND":
		s.handleSend(client, cmd)
	case "STORE":
		s.handleStore(client, cmd)
	case "LOGOUT":
		s.handleLogout(client, cmd)
	default:
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "Unknown command",
		})
	}
}

func (s *PipecServer) handleLogin(client *Client, cmd Command) {
	if client.state != StateNotAuthenticated {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "Already authenticated",
		})
		return
	}

	username, ok := cmd.Params["username"].(string)
	if !ok {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "Username required",
		})
		return
	}

	password, ok := cmd.Params["password"].(string)
	if !ok {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "Password required",
		})
		return
	}

	// Authenticate user
	var user User
	err := s.db.QueryRow("SELECT id, username, email, password_hash FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)

	if err != nil {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "Authentication failed",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "Authentication failed",
		})
		return
	}

	client.state = StateAuthenticated
	client.userID = user.ID
	client.username = user.Username

	s.sendResponse(client, Response{
		ID:      cmd.ID,
		Status:  "OK",
		Message: "Login successful",
	})
}

func (s *PipecServer) handleList(client *Client, cmd Command) {
	if client.state == StateNotAuthenticated {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "Not authenticated",
		})
		return
	}

	rows, err := s.db.Query("SELECT id, name, created_at FROM mailboxes WHERE user_id = $1", client.userID)
	if err != nil {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "Failed to list mailboxes",
		})
		return
	}
	defer rows.Close()

	var mailboxes []Mailbox
	for rows.Next() {
		var mbox Mailbox
		err := rows.Scan(&mbox.ID, &mbox.Name, &mbox.CreatedAt)
		if err != nil {
			continue
		}
		mbox.UserID = client.userID
		mailboxes = append(mailboxes, mbox)
	}

	s.sendResponse(client, Response{
		ID:      cmd.ID,
		Status:  "OK",
		Message: "Mailboxes listed",
		Data:    mailboxes,
	})
}

func (s *PipecServer) handleSelect(client *Client, cmd Command) {
	if client.state == StateNotAuthenticated {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "Not authenticated",
		})
		return
	}

	mailbox, ok := cmd.Params["mailbox"].(string)
	if !ok {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "Mailbox name required",
		})
		return
	}

	// Get mailbox info
	var mboxID int
	var messageCount, recentCount, unseenCount int

	err := s.db.QueryRow("SELECT id FROM mailboxes WHERE user_id = $1 AND name = $2",
		client.userID, mailbox).Scan(&mboxID)
	if err != nil {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "Mailbox not found",
		})
		return
	}

	// Count messages
	s.db.QueryRow("SELECT COUNT(*) FROM messages WHERE mailbox_id = $1", mboxID).Scan(&messageCount)

	// Count recent (last 24 hours)
	s.db.QueryRow("SELECT COUNT(*) FROM messages WHERE mailbox_id = $1 AND created_at > $2",
		mboxID, time.Now().Add(-24*time.Hour)).Scan(&recentCount)

	// Count unseen
	s.db.QueryRow("SELECT COUNT(*) FROM messages WHERE mailbox_id = $1 AND NOT ($2 = ANY(flags))",
		mboxID, "SEEN").Scan(&unseenCount)

	client.state = StateSelected
	client.selectedMbox = mailbox
	client.selectedMboxID = mboxID

	info := MailboxInfo{
		Name:     mailbox,
		Messages: messageCount,
		Recent:   recentCount,
		Unseen:   unseenCount,
	}

	s.sendResponse(client, Response{
		ID:      cmd.ID,
		Status:  "OK",
		Message: "Mailbox selected",
		Data:    info,
	})
}

func (s *PipecServer) handleFetch(client *Client, cmd Command) {
	if client.state != StateSelected {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "No mailbox selected",
		})
		return
	}

	sequence, ok := cmd.Params["sequence"].(string)
	if !ok {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "Sequence required",
		})
		return
	}

	itemsInterface, ok := cmd.Params["items"].([]interface{})
	if !ok {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "Items required",
		})
		return
	}

	var items []string
	for _, item := range itemsInterface {
		if str, ok := item.(string); ok {
			items = append(items, str)
		}
	}

	// TODO: Use items to filter which fields to return (e.g., BODY, HEADERS, FLAGS)
	_ = items // Acknowledge that items is currently unused

	// Parse sequence (simple implementation for ranges like "1:10")
	var query string
	var args []interface{}

	if strings.Contains(sequence, ":") {
		parts := strings.Split(sequence, ":")
		if len(parts) == 2 {
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			query = `SELECT id, uid, from_addr, to_addrs, subject, body, flags, headers, created_at 
					 FROM messages WHERE mailbox_id = $1 ORDER BY id LIMIT $2 OFFSET $3`
			args = []interface{}{client.selectedMboxID, end - start + 1, start - 1}
		}
	} else {
		// Single message
		msgNum, _ := strconv.Atoi(sequence)
		query = `SELECT id, uid, from_addr, to_addrs, subject, body, flags, headers, created_at 
				 FROM messages WHERE mailbox_id = $1 ORDER BY id LIMIT 1 OFFSET $2`
		args = []interface{}{client.selectedMboxID, msgNum - 1}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "Failed to fetch messages",
		})
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var toAddrsJSON, flagsJSON, headersJSON []byte

		err := rows.Scan(&msg.ID, &msg.UID, &msg.FromAddr, &toAddrsJSON,
			&msg.Subject, &msg.Body, &flagsJSON, &headersJSON, &msg.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(toAddrsJSON, &msg.ToAddrs)
		json.Unmarshal(flagsJSON, &msg.Flags)
		json.Unmarshal(headersJSON, &msg.Headers)

		msg.MailboxID = client.selectedMboxID
		messages = append(messages, msg)
	}

	s.sendResponse(client, Response{
		ID:      cmd.ID,
		Status:  "OK",
		Message: "Messages fetched",
		Data:    messages,
	})
}

func (s *PipecServer) handleSend(client *Client, cmd Command) {
	if client.state == StateNotAuthenticated {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "Not authenticated",
		})
		return
	}

	from, ok := cmd.Params["from"].(string)
	if !ok {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "From address required",
		})
		return
	}

	toInterface, ok := cmd.Params["to"].([]interface{})
	if !ok {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "To addresses required",
		})
		return
	}

	var to []string
	for _, addr := range toInterface {
		if str, ok := addr.(string); ok {
			to = append(to, str)
		}
	}

	subject, _ := cmd.Params["subject"].(string)
	body, _ := cmd.Params["body"].(string)

	// Get user's SENT mailbox
	var sentMboxID int
	err := s.db.QueryRow("SELECT id FROM mailboxes WHERE user_id = $1 AND name = $2",
		client.userID, "SENT").Scan(&sentMboxID)

	if err != nil {
		// Create SENT mailbox if it doesn't exist
		err = s.db.QueryRow("INSERT INTO mailboxes (user_id, name, created_at) VALUES ($1, $2, $3) RETURNING id",
			client.userID, "SENT", time.Now()).Scan(&sentMboxID)
		if err != nil {
			s.sendResponse(client, Response{
				ID:      cmd.ID,
				Status:  "NO",
				Message: "Failed to create SENT mailbox",
			})
			return
		}
	}

	// Store message in SENT folder
	uid := uuid.New().String()
	toJSON, _ := json.Marshal(to)
	flagsJSON, _ := json.Marshal([]string{"SENT"})
	headers := map[string]interface{}{
		"message-id": fmt.Sprintf("<%s@%s>", uid, "pipec.local"),
		"date":       time.Now().Format(time.RFC3339),
	}
	headersJSON, _ := json.Marshal(headers)

	_, err = s.db.Exec(`INSERT INTO messages (mailbox_id, uid, from_addr, to_addrs, subject, body, flags, headers, created_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		sentMboxID, uid, from, toJSON, subject, body, flagsJSON, headersJSON, time.Now())

	if err != nil {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "Failed to store message",
		})
		return
	}

	// TODO: Here you would typically forward the message via PIPE protocol to other servers

	s.sendResponse(client, Response{
		ID:      cmd.ID,
		Status:  "OK",
		Message: "Message sent and stored",
	})
}

func (s *PipecServer) handleStore(client *Client, cmd Command) {
	if client.state != StateSelected {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "No mailbox selected",
		})
		return
	}

	sequence, ok := cmd.Params["sequence"].(string)
	if !ok {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "Sequence required",
		})
		return
	}

	flagsInterface, ok := cmd.Params["flags"].([]interface{})
	if !ok {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "BAD",
			Message: "Flags required",
		})
		return
	}

	var flags []string
	for _, flag := range flagsInterface {
		if str, ok := flag.(string); ok {
			flags = append(flags, str)
		}
	}

	flagsJSON, _ := json.Marshal(flags)

	// Simple sequence parsing (single message for now)
	msgNum, _ := strconv.Atoi(sequence)

	_, err := s.db.Exec(`UPDATE messages SET flags = $1 
						WHERE mailbox_id = $2 AND id = (
							SELECT id FROM messages WHERE mailbox_id = $2 ORDER BY id LIMIT 1 OFFSET $3
						)`,
		flagsJSON, client.selectedMboxID, msgNum-1)

	if err != nil {
		s.sendResponse(client, Response{
			ID:      cmd.ID,
			Status:  "NO",
			Message: "Failed to store flags",
		})
		return
	}

	s.sendResponse(client, Response{
		ID:      cmd.ID,
		Status:  "OK",
		Message: "Flags stored",
	})
}

func (s *PipecServer) handleLogout(client *Client, cmd Command) {
	client.state = StateLogout
	s.sendResponse(client, Response{
		ID:      cmd.ID,
		Status:  "OK",
		Message: "Logout successful",
	})
}

func (s *PipecServer) sendResponse(client *Client, resp Response) {
	data, _ := json.Marshal(resp)
	client.conn.Write(append(data, '\n'))
}
