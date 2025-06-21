package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EmailMessage represents a stored email message
type EmailMessage struct {
	ID       string            `json:"id"`
	From     string            `json:"from"`
	To       []string          `json:"to"`
	Subject  string            `json:"subject"`
	Body     string            `json:"body"`
	Headers  map[string]string `json:"headers"`
	Received time.Time         `json:"received"`
	Status   string            `json:"status"`
}

// FileStorage implements file-based storage
type FileStorage struct {
	basePath string
}

// NewFileStorage creates a new file storage instance
func NewFileStorage(basePath string) *FileStorage {
	return &FileStorage{
		basePath: basePath,
	}
}

// Initialize creates necessary directories
func (fs *FileStorage) Initialize() error {
	dirs := []string{
		filepath.Join(fs.basePath, "inbox"),
		filepath.Join(fs.basePath, "outbox"),
		filepath.Join(fs.basePath, "sent"),
		filepath.Join(fs.basePath, "queue"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// StoreIncoming stores an incoming email message
func (fs *FileStorage) StoreIncoming(msg *EmailMessage) error {
	msg.Received = time.Now()
	msg.Status = "received"

	filename := fmt.Sprintf("%d_%s.json", time.Now().UnixNano(), msg.ID)
	filepath := filepath.Join(fs.basePath, "inbox", filename)

	return fs.writeMessage(filepath, msg)
}

// StoreOutgoing stores an outgoing email message
func (fs *FileStorage) StoreOutgoing(msg *EmailMessage) error {
	msg.Status = "sending"

	filename := fmt.Sprintf("%d_%s.json", time.Now().UnixNano(), msg.ID)
	filepath := filepath.Join(fs.basePath, "outbox", filename)

	return fs.writeMessage(filepath, msg)
}

// MarkSent moves a message from outbox to sent folder
func (fs *FileStorage) MarkSent(messageID string) error {
	// Find the message in outbox
	outboxPath := filepath.Join(fs.basePath, "outbox")
	sentPath := filepath.Join(fs.basePath, "sent")

	var foundFile string
	err := filepath.WalkDir(outboxPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.Contains(d.Name(), messageID) {
			foundFile = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		return err
	}

	if foundFile == "" {
		return fmt.Errorf("message %s not found in outbox", messageID)
	}

	// Read, update status, and move
	msg, err := fs.readMessage(foundFile)
	if err != nil {
		return err
	}

	msg.Status = "sent"

	newFilename := filepath.Base(foundFile)
	newPath := filepath.Join(sentPath, newFilename)

	if err := fs.writeMessage(newPath, msg); err != nil {
		return err
	}

	return os.Remove(foundFile)
}

// GetInboxMessages returns all messages in inbox
func (fs *FileStorage) GetInboxMessages() ([]*EmailMessage, error) {
	return fs.getMessagesFromDir(filepath.Join(fs.basePath, "inbox"))
}

// GetOutboxMessages returns all messages in outbox
func (fs *FileStorage) GetOutboxMessages() ([]*EmailMessage, error) {
	return fs.getMessagesFromDir(filepath.Join(fs.basePath, "outbox"))
}

// writeMessage writes a message to file
func (fs *FileStorage) writeMessage(filepath string, msg *EmailMessage) error {
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// readMessage reads a message from file
func (fs *FileStorage) readMessage(filepath string) (*EmailMessage, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var msg EmailMessage
	err = json.Unmarshal(data, &msg)
	return &msg, err
}

// getMessagesFromDir returns all messages from a directory
func (fs *FileStorage) getMessagesFromDir(dir string) ([]*EmailMessage, error) {
	var messages []*EmailMessage

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			msg, err := fs.readMessage(path)
			if err != nil {
				return err
			}
			messages = append(messages, msg)
		}

		return nil
	})

	return messages, err
}
