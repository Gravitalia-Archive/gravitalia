package grpc

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/Gravitalia/gravitalia/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TagImage provides a way to obtain tag of an images
func TagImage(model int32, image []byte) (string, error) {
	// Set up a connection to the server
	conn, err := grpc.Dial(os.Getenv("TORRESIX_ADDRESSS"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}

	defer conn.Close()
	c := proto.NewTorreClient(conn)

	// Contact the server and print out its response
	// If no response in 5 seconds, cancel it
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Make request, and start predict label
	r, err := c.TorrePredict(ctx, &proto.TorreRequest{
		Model: model,
		Data:  image,
	})

	if err != nil {
		return "", err
	}

	// Check if error is sent
	if r.GetError() {
		return "", errors.New(r.GetMessage())
	}

	return r.GetMessage(), nil
}
