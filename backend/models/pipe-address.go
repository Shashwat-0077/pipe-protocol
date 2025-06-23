package models

// PipeAddress represents a parsed PIPE email address
type PipeAddress struct {
	Username string `json:"username"` // Username part
	Domain   string `json:"domain"`   // Domain part
	Raw      string `json:"raw"`      // Full address (username|domain)
}
