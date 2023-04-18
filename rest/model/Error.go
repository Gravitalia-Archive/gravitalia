package model

// RequestError represents the structure of the response, in case of error
type RequestError struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}
