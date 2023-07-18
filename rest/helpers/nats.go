package helpers

import (
	"log"
	"os"

	"github.com/nats-io/nats.go"
)

var Nats *nats.Conn

// InitNATS starts a new NATS instance
func InitNATS() {
	connection, err := nats.Connect(os.Getenv("NATS_URL"))

	if err != nil {
		log.Printf("Cannot connect to %v: %v", os.Getenv("NATS_URL"), err)
	}

	Nats = connection
}

// Publish allows publishing message on NATS
func Publish(subject string, message []byte) {
	err := Nats.Publish(subject, message)

	if err != nil {
		log.Printf("(Publish) Failed to send message to %v, got error: %v", subject, err)
	}
}
