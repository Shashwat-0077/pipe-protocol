package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"pipe-protocol/pkg/protocol"
	"strings"
	"time"
)

// Client represents a PIPE client for sending messages
type Client struct {
	serverName string
	timeout    time.Duration
}

// NewClient creates a new PIPE client
func NewClient(serverName string) *Client {
	return &Client{
		serverName: serverName,
		timeout:    30 * time.Second,
	}
}

// SendMessage sends a message to a remote server
func (c *Client) SendMessage(host string, port int, from string, to []string, subject, body string, headers map[string]string) error {
	// Connect to remote server
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, c.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(c.timeout))

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Send handshake
	handshake := protocol.NewHandshake(c.serverName)
	if err := c.sendMessage(writer, handshake); err != nil {
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	// Read handshake response
	response, err := c.readResponse(reader)
	if err != nil {
		return fmt.Errorf("failed to read handshake response: %w", err)
	}

	if response.Status != protocol.StatusOK {
		return fmt.Errorf("handshake failed: %s", response.Message)
	}

	// Send message
	send := protocol.NewSend(from, to, subject, body, headers)
	if err := c.sendMessage(writer, send); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Read send response
	response, err = c.readResponse(reader)
	if err != nil {
		return fmt.Errorf("failed to read send response: %w", err)
	}

	if response.Status == protocol.StatusError {
		return fmt.Errorf("message rejected: %s", response.Message)
	}

	fmt.Printf("Message sent successfully: %s\n", response.Message)
	if response.Details != nil {
		if len(response.Details.Accepted) > 0 {
			fmt.Printf("Accepted: %v\n", response.Details.Accepted)
		}
		if len(response.Details.Rejected) > 0 {
			fmt.Printf("Rejected: %v (Reason: %s)\n", response.Details.Rejected, response.Details.Reason)
		}
	}

	return nil
}

// sendMessage sends a message over the connection
func (c *Client) sendMessage(writer *bufio.Writer, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = writer.Write(append(data, '\n'))
	if err != nil {
		return err
	}

	return writer.Flush()
}

// readResponse reads a response from the connection
func (c *Client) readResponse(reader *bufio.Reader) (*protocol.Response, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	msg, err := protocol.ParseMessage([]byte(line))
	if err != nil {
		return nil, err
	}

	response, ok := msg.(*protocol.Response)
	if !ok {
		return nil, fmt.Errorf("expected response message, got %T", msg)
	}

	return response, nil
}
