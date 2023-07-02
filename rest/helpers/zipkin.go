package helpers

import (
	"log"
	"net/http"
	"os"

	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"
	logreporter "github.com/openzipkin/zipkin-go/reporter/log"
)

func InitTracer() (*zipkinhttp.Client, func(http.Handler) http.Handler) {
	// set up a span reporter
	reporter := logreporter.NewReporter(log.New(os.Stderr, "", log.LstdFlags))
	defer func() {
		_ = reporter.Close()
	}()

	// create our local service endpoint
	endpoint, err := zipkin.NewEndpoint("gravitaliaRest", os.Getenv("ZIPKIN_ADDRESS"))
	if err != nil {
		log.Printf("unable to create local endpoint: %+v\n", err)
	}

	// initialize our tracer
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Printf("unable to create tracer: %+v\n", err)
	}

	// create global zipkin http server middleware
	serverMiddleware := zipkinhttp.NewServerMiddleware(
		tracer, zipkinhttp.TagResponseSize(true),
	)

	// create global zipkin traced http client
	client, err := zipkinhttp.NewClient(tracer, zipkinhttp.ClientTrace(true))
	if err != nil {
		log.Printf("unable to create client: %+v\n", err)
	}

	return client, serverMiddleware
}
