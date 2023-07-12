package model

// Message represents how NATS publish message
// should be
type Message struct {
	Type      string `json:"type"`
	From      string `json:"from"`
	To        string `json:"string"`
	Important bool   `json:"important"`
}
