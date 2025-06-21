package handlers

import (
	"database/sql"
	"pipec-backend/models"
	"pipec-backend/types"
	"time"
)

func handleSelect(db *sql.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State == types.StateNotAuthenticated {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Not authenticated"})
		return
	}

	mailbox, ok := cmd.Params["mailbox"].(string)
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Mailbox name required"})
		return
	}

	var mboxID int
	var messageCount, recentCount, unseenCount int

	err := db.QueryRow("SELECT id FROM mailboxes WHERE user_id = $1 AND name = $2", client.UserID, mailbox).
		Scan(&mboxID)
	if err != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Mailbox not found"})
		return
	}
	db.QueryRow("SELECT COUNT(*) FROM messages WHERE mailbox_id = $1", mboxID).Scan(&messageCount)
	db.QueryRow("SELECT COUNT(*) FROM messages WHERE mailbox_id = $1 AND created_at > $2", mboxID, time.Now().Add(-24*time.Hour)).Scan(&recentCount)
	db.QueryRow("SELECT COUNT(*) FROM messages WHERE mailbox_id = $1 AND NOT ($2 = ANY(flags))", mboxID, "SEEN").Scan(&unseenCount)

	client.State = types.StateSelected
	client.SelectedMbox = mailbox
	client.SelectedMboxID = mboxID

	info := models.MailboxInfo{
		Name:     mailbox,
		Messages: messageCount,
		Recent:   recentCount,
		Unseen:   unseenCount,
	}

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Mailbox selected", Data: info})
}
