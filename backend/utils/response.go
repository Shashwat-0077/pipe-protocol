package utils // or types if preferred

import (
	"encoding/json"
	"net"
	"pipec-backend/models"
)

func SendResponse(conn net.Conn, resp models.Response) {
	data, _ := json.Marshal(resp)
	conn.Write(append(data, '\n'))
}
