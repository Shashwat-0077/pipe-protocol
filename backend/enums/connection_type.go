package enums

// ConnectionType represents the type of connection used for commands
type ConnectionType string

const (
	ConnectionTypeClient ConnectionType = "client"
	ConnectionTypeServer ConnectionType = "server"
)

// IsValid checks if the ConnectionType is valid
func (ct ConnectionType) IsValid() bool {
	switch ct {
	case ConnectionTypeClient, ConnectionTypeServer:
		return true
	default:
		return false
	}
}
