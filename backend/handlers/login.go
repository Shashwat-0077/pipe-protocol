package handlers

import (
	"pipec-backend/models"
	"pipec-backend/types"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func handleLogin(db *gorm.DB, client *types.Client, cmd *models.Command, send SendResponseFunc) {
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
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		send(client, models.Response{ID: cmd.ID, Status: "NO", Message: "Authentication failed"})
		return
	}

	client.State = types.StateAuthenticated
	client.UserID = int(user.ID)
	client.Username = user.Username

	send(client, models.Response{ID: cmd.ID, Status: "OK", Message: "Login successful"})
}
