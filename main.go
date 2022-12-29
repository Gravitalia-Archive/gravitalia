package main

import (
	"net/http"
	"os"

	"log"

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

	log.Println("Server is starting on port", os.Getenv("PORT"))
	// Create web server
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
