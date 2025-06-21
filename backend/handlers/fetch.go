package handlers

import (
	"database/sql"
	"encoding/json"
	"pipec-backend/models"
	"pipec-backend/types"
	"strconv"
	"strings"
)

func handleFetch(db *sql.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State != types.StateSelected {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "No mailbox selected"})
		return
	}

	sequence, ok := cmd.Params["sequence"].(string)
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Sequence required"})
		return
	}

	_, ok = cmd.Params["items"].([]interface{})

	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Items required"})
		return
	}

	var query string
	var args []interface{}

	if strings.Contains(sequence, ":") {
		parts := strings.Split(sequence, ":")
		if len(parts) == 2 {
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			query = `SELECT id, uid, from_addr, to_addrs, subject, body, flags, headers, created_at FROM messages WHERE mailbox_id = $1 ORDER BY id LIMIT $2 OFFSET $3`
			args = []interface{}{client.SelectedMboxID, end - start + 1, start - 1}
		}
	} else {
		msgNum, _ := strconv.Atoi(sequence)
		query = `SELECT id, uid, from_addr, to_addrs, subject, body, flags, headers, created_at FROM messages WHERE mailbox_id = $1 ORDER BY id LIMIT 1 OFFSET $2`
		args = []interface{}{client.SelectedMboxID, msgNum - 1}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to fetch messages"})
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var toAddrsJSON, flagsJSON, headersJSON []byte
		err := rows.Scan(&msg.ID, &msg.UID, &msg.FromAddr, &toAddrsJSON, &msg.Subject, &msg.Body, &flagsJSON, &headersJSON, &msg.CreatedAt)
		if err != nil {
			continue
		}
		json.Unmarshal(toAddrsJSON, &msg.ToAddrs)
		json.Unmarshal(flagsJSON, &msg.Flags)
		json.Unmarshal(headersJSON, &msg.Headers)
		msg.MailboxID = client.SelectedMboxID
		messages = append(messages, msg)
	}

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Messages fetched", Data: messages})
}
