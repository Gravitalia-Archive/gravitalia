package model

type AddBody struct {
	Content string `json:"content"`
	ReplyTo string `json:"reply,omitempty"`
}
