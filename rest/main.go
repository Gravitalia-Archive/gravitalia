package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	route "github.com/Gravitalia/gravitalia/router"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting server...")

	// Get key-value in .env file
	godotenv.Load()

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

	// Add tracer
	client, serverMiddleware := helpers.InitTracer()

	router.HandleFunc("/", route.Index)
	router.HandleFunc("/callback", route.OAuth(client))
	router.HandleFunc("/users/", route.UserHandler)
	router.HandleFunc("/relation/", route.RelationHandler)
	router.HandleFunc("/posts/", route.PostHandler)
	router.HandleFunc("/comment/", route.Handler)
	router.HandleFunc("/list/", route.ListHandler)
	router.HandleFunc("/request/", route.AcceptOrDecline)
	router.HandleFunc("/account/deletion", route.DeleteUser(client))
	router.HandleFunc("/account/suspend", route.Suspend)
	router.HandleFunc("/account/data", route.GetData)
	router.Handle("/metrics", promhttp.HandlerFor(helpers.GetRegistery(), promhttp.HandlerOpts{}))

	// Init every helpers function and database variables
	helpers.Init()
	database.Init()
	helpers.InitNATS()

	log.Println("Server is starting on port", os.Getenv("PORT"))

	// Create web server
	server := &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		ReadHeaderTimeout: 3 * time.Second,
	}
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middleware(serverMiddleware(router)).ServeHTTP(w, r)
	})

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
