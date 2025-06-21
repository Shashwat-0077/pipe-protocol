// File: handlers/login.go
package handlers

import (
	"database/sql"
	"pipec-backend/models"
	"pipec-backend/types"

	"golang.org/x/crypto/bcrypt"
)

func handleLogin(db *sql.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
	if client.State != types.StateNotAuthenticated {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Already authenticated"})
		return
	}

	username, ok := cmd.Params["username"].(string)
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Username required"})
		return
	}
	password, ok := cmd.Params["password"].(string)
	if !ok {
		send(client, models.Response{ID: cmd.ID, Status: "BAD", Message: "Password required"})
		return
	}

	var user models.User
	err := db.QueryRow("SELECT id, username, email, password_hash FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Authentication failed"})
		return
	}

	client.State = types.StateAuthenticated
	client.UserID = user.ID
	client.Username = user.Username

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Login successful"})
}
