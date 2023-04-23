package main

import (
	"net/http"
	"os"
	"time"

	"log"

	"github.com/Gravitalia/recommendation/database"
	"github.com/Gravitalia/recommendation/router"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	// Get key-value in .env file
	godotenv.Load()

	// Start a new cron job
	c := cron.New()
	c.AddFunc("@hourly", func() { // switch to @daily or @weekly when Gravitalia grows
		log.Println("Starting PageRank and Community Detection...")
		_, err := database.PageRank()
		if err != nil {
			log.Panicf("PageRank did not work as expected")
		}
		_, err = database.CommunityDetection()
		if err != nil {
			log.Panicf("Community Detection did not work as expected")
		}
	})
	c.Start()

	// Create routes
	http.HandleFunc("/", router.Index)
	http.HandleFunc("/for_you_feed", router.Get)

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
