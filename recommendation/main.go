package main

import (
	"net/http"
	"os"
	"time"

	"log"

	"github.com/Gravitalia/recommendation/database"
	"github.com/Gravitalia/recommendation/helpers"
	route "github.com/Gravitalia/recommendation/router"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Init database
	database.Init()

	// Create a middleware to count requests
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
			} else {
				start := time.Now()
				helpers.IncrementRequests()

				next.ServeHTTP(w, r)

				helpers.ObserveRequestDuration(time.Since(start).Seconds())
			}
		})
	}

	// Create routes
	router := http.NewServeMux()
	router.HandleFunc("/", route.Index)
	router.HandleFunc("/recommendation/for_you_feed", route.Handler)
	router.Handle("/metrics", promhttp.HandlerFor(helpers.GetRegistery(), promhttp.HandlerOpts{}))

	log.Println("Server is starting on port", os.Getenv("RECOMMENDATION_PORT"))

	// Create web server
	server := &http.Server{
		Addr:              ":" + os.Getenv("RECOMMENDATION_PORT"),
		ReadHeaderTimeout: 3 * time.Second,
	}
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middleware(router).ServeHTTP(w, r)
	})

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
