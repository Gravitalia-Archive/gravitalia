package helpers

import (
	"log"
	"os"

	"github.com/nats-io/nats.go"
)

var Nats *nats.Conn

// InitNATS starts a new NATS instance
func InitNATS() {
	Nats, _ = nats.Connect(os.Getenv("NATS_URL"))
}

// Publish allows publishing message on NATS
func Publish(subject string, message []byte) {
	err := Nats.Publish(subject, message)

	if err != nil {
		log.Printf("(Publish) Failed to send message to %v, got error: %v", subject, err)
	}
}
