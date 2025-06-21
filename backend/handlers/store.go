package handlers

import (
	"database/sql"
	"encoding/json"
	"pipec-backend/models"
	"pipec-backend/types"
	"strconv"
)

func handleStore(db *sql.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State != types.StateSelected {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "No mailbox selected"})
		return
	}

	sequence, ok := cmd.Params["sequence"].(string)
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Sequence required"})
		return
	}

	flagsInterface, ok := cmd.Params["flags"].([]interface{})
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Flags required"})
		return
	}

	var flags []string
	for _, flag := range flagsInterface {
		if str, ok := flag.(string); ok {
			flags = append(flags, str)
		}
	}

	flagsJSON, _ := json.Marshal(flags)
	msgNum, _ := strconv.Atoi(sequence)

	_, err := db.Exec(`UPDATE messages SET flags = $1 WHERE mailbox_id = $2 AND id = (
		SELECT id FROM messages WHERE mailbox_id = $2 ORDER BY id LIMIT 1 OFFSET $3)`, flagsJSON, client.SelectedMboxID, msgNum-1)
	if err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to store flags"})
		return
	}

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Flags stored"})
}
