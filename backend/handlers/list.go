package handlers

import (
	"database/sql"
	"pipec-backend/models"
	"pipec-backend/types"
)

func handleList(db *sql.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State == types.StateNotAuthenticated {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Not authenticated"})
		return
	}

	rows, err := db.Query("SELECT id, name, created_at FROM mailboxes WHERE user_id = $1", client.UserID)
	if err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Failed to list mailboxes"})
		return
	}
	defer rows.Close()

	var mailboxes []models.Mailbox
	for rows.Next() {
		var mbox models.Mailbox
		if err := rows.Scan(&mbox.ID, &mbox.Name, &mbox.CreatedAt); err != nil {
			continue
		}
		mbox.UserID = client.UserID
		mailboxes = append(mailboxes, mbox)
	}

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Mailboxes listed", Data: mailboxes})
}
