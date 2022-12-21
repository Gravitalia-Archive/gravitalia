package model

// AuthaUser structure represents user that tries to connect
type AuthaUser struct {
	Username string `json:"username"`
	Vanity   string `json:"vanity"`
}
