package main

import (
	"net/http"
	"os"
	"time"

	"log"

	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/router"
	"github.com/joho/godotenv"
)

func main() {
	// Get key-value in .env file
	godotenv.Load()

	// Create routes
	http.HandleFunc("/", router.Index)
	http.HandleFunc("/callback", router.OAuth)
	http.HandleFunc("/v1/new", router.New)
	http.HandleFunc("/users/", router.Users)
	http.HandleFunc("/relation/", router.Relation)
	http.HandleFunc("/posts/", router.GetPost)

	// Init every helpers function
	helpers.Init()

	log.Println("Server is starting on port", os.Getenv("PORT"))
	// Create web server
	server := &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		ReadHeaderTimeout: 3 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
