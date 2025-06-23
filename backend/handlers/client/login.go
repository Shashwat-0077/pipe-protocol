package handlers

import (
	"fmt"
	"pipec-backend/enums"
	"pipec-backend/models"
	"pipec-backend/repository"
	"pipec-backend/types"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginParams struct {
	username string
	password string
}

func parseLoginParams(cmd *models.Command, client *types.Client, send types.SendResponseFunc) (*LoginParams, bool) {
	params := &LoginParams{}

	username, ok := cmd.Params["username"].(string)
	if !ok {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusBAD, Message: "Username required"})
		return nil, false
	}

	password, ok := cmd.Params["password"].(string)
	if !ok {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusBAD, Message: "Password required"})
		return nil, false
	}

	params.username = username
	params.password = password
	return params, true
}

func handleLogin(db *gorm.DB, client *types.Client, cmd *models.Command, send types.SendResponseFunc) {
	if isAuthenticated(client) {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusBAD, Message: "Already authenticated"})
		return
	}

	params, ok := parseLoginParams(cmd, client, send)
	if !ok {
		return
	}

	fmt.Printf("\033[36mAttempting login for user: %s\033[0m\n", params.username)

	user, err := repository.GetUserByUsername(db, params.username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.password)) != nil {
		send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusNO, Message: "Authentication failed"})
		return
	}

	client.State = types.StateAuthenticated
	client.User = user

	send(client.Conn, models.Response{ID: cmd.ID, Status: enums.StatusOK, Message: "Login successful"})
}
