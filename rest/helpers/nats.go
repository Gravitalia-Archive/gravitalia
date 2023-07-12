package helpers

import (
	"os"

	"github.com/nats-io/nats.go"
)

var Nats *nats.Conn

func InitNATS() {
	Nats, _ = nats.Connect(os.Getenv("NATS_URL"))
}

func Publish(subject string, message []byte) {
	Nats.Publish(subject, message)
}
