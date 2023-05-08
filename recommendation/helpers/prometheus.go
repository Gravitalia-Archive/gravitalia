package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Tracks the number of HTTP requests.",
	})

	requestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Tracks the latencies for HTTP requests.",
		Buckets: prometheus.DefBuckets,
	})
)

// GetRegistery is used to get prometheus
// saved data
func GetRegistery() *prometheus.Registry {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		requestsTotal,
		requestDuration,
	)

	return registry
}

// IncrementRequests allows to increment
// the number of total requests
func IncrementRequests() {
	requestsTotal.Inc()
}

// ObserveRequestDuration allows to create a
// new record of time duration
func ObserveRequestDuration(time float64) {
	requestDuration.Observe(time)
}
