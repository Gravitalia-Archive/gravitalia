package model

type Request struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type AuthaUser struct {
	Username string `json:"username"`
	Vanity   string `json:"vanity"`
}
