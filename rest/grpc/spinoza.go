package grpc

import (
	"context"
	"os"
	"time"

	"github.com/Gravitalia/gravitalia/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UploadImage allows to transfer image as bytes
// into Spinoza server to upload it to image provider
func UploadImage(image []byte) (string, error) {
	// Set up a connection to the server
	conn, err := grpc.Dial(os.Getenv("SPINOZA_ADDRESS"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}

	defer conn.Close()
	c := proto.NewSpinozaClient(conn)

	// Contact the server and print out its response
	// If no response in 20 seconds, cancel it
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// Make request
	r, err := c.Upload(ctx, &proto.UploadRequest{
		Data: image,
		//Width: 3840, // 1920 for FHD
	})

	if err != nil {
		return "", err
	}

	return r.GetMessage(), nil
}

// DeleteImage allows to delete an image with its
// hash. Returns the error message (may be empty)
func DeleteImage(hash string) (string, error) {
	// Set up a connection to the server
	conn, err := grpc.Dial(os.Getenv("SPINOZA_ADDRESS"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}

	defer conn.Close()
	c := proto.NewSpinozaClient(conn)

	// Contact the server and print out its response
	// If no response in 2 seconds, cancel it
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	// Make request
	r, err := c.Delete(ctx, &proto.DeleteRequest{
		Hash: hash,
	})

	if err != nil {
		return "", err
	}

	return r.GetMessage(), nil
}
