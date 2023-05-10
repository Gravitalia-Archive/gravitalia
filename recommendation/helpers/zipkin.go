package helpers

import (
	"log"
	"os"

	"github.com/openzipkin/zipkin-go"
	logreporter "github.com/openzipkin/zipkin-go/reporter/log"
)

// NewTracer allows to create a Zipkin tracer
func NewTracer(port string) (*zipkin.Tracer, error) {
	reporter := logreporter.NewReporter(log.New(os.Stderr, "", log.LstdFlags))
	defer reporter.Close()

	endpoint, err := zipkin.NewEndpoint("gravitaliaRecommendation", "localhost:"+port)
	if err != nil {
		return nil, err
	}

	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		return nil, err
	}

	return tracer, nil
}
