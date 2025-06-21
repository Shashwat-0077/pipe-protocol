package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"pipec-backend/models"
	"pipec-backend/types"
	"time"

	"github.com/google/uuid"
)

func handleSend(db *sql.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State == types.StateNotAuthenticated {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Not authenticated"})
		return
	}

	from, ok := cmd.Params["from"].(string)
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "From address required"})
		return
	}
	toInterface, ok := cmd.Params["to"].([]interface{})
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "To addresses required"})
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

	var sentMboxID int
	err := db.QueryRow("SELECT id FROM mailboxes WHERE user_id = $1 AND name = $2", client.UserID, "SENT").Scan(&sentMboxID)
	if err != nil {
		err = db.QueryRow("INSERT INTO mailboxes (user_id, name, created_at) VALUES ($1, $2, $3) RETURNING id", client.UserID, "SENT", time.Now()).Scan(&sentMboxID)
		if err != nil {
			send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to create SENT mailbox"})
			return
		}
	}

	uid := uuid.New().String()
	toJSON, _ := json.Marshal(to)
	flagsJSON, _ := json.Marshal([]string{"SENT"})
	headers := map[string]interface{}{
		"message-id": fmt.Sprintf("<%s@%s>", uid, "pipec.local"),
		"date":       time.Now().Format(time.RFC3339),
	}
	headersJSON, _ := json.Marshal(headers)

	_, err = db.Exec(`INSERT INTO messages (mailbox_id, uid, from_addr, to_addrs, subject, body, flags, headers, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, sentMboxID, uid, from, toJSON, subject, body, flagsJSON, headersJSON, time.Now())
	if err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to store message"})
		return
	}

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Message sent and stored"})
}
