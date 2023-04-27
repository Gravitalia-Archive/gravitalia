package main

import (
	"net/http"
	"os"
	"time"

	"log"

	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/router"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Get key-value in .env file
	godotenv.Load()

	// Prometheus
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	// Create routes
	http.HandleFunc("/", router.Index)
	http.HandleFunc("/callback", router.OAuth)
	http.HandleFunc("/v1/new", router.New)
	http.HandleFunc("/users/", router.Users)
	http.HandleFunc("/relation/", router.Relation)
	http.HandleFunc("/posts/", router.Post)
	http.Handle(
		"/metrics",
		helpers.New(
			registry, nil).
			WrapHandler("/metrics", promhttp.HandlerFor(
				registry,
				promhttp.HandlerOpts{}),
			))

	// Init every helpers function
	helpers.Init()

	log.Println("Server is starting on port", os.Getenv("PORT"))

	// Create web server
	server := &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		ReadHeaderTimeout: 3 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
