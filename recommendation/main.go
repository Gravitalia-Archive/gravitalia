package main

import (
	"net/http"
	"os"
	"time"

	"log"

	"github.com/Gravitalia/recommendation/router"
	"github.com/joho/godotenv"
)

func main() {
	// Get key-value in .env file
	godotenv.Load()

	// Create routes
	http.HandleFunc("/", router.Index)

	log.Println("Server is starting on port", os.Getenv("RECOMMENDATION_PORT"))

	// Create web server
	server := &http.Server{
		Addr:              ":" + os.Getenv("RECOMMENDATION_PORT"),
		ReadHeaderTimeout: 3 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
