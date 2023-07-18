package model

// Message represents how NATS publish message
// should be
type Message struct {
	Type string `json:"type"`
	// Author vanity (such as realhinome)
	From string `json:"from"`
	// Must be User vanity or post ID
	To string `json:"to"`
	// Set true to send push notification
	Important bool `json:"important"`
}
