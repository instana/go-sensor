package instana_test

import (
	"context"
	"log"
	"net/http"

	instana "github.com/instana/go-sensor"
)

// This example shows how to instrument an HTTP client with Instana tracing
func Example_roundTripper() {
	sensor := instana.NewSensor("my-http-client")

	// Wrap the original http.Client transport with instana.RoundTripper().
	// The http.DefaultTransport will be used if there was no transport provided.
	client := &http.Client{
		Transport: instana.RoundTripper(sensor, nil),
	}

	// Use your instrumented http.Client to propagate tracing context with the request
	_, err := client.Get("https://www.instana.com")
	if err != nil {
		log.Fatalf("failed to GET https://www.instana.com: %s", err)
	}

	// To propagate the existing trace with request, make sure that current span is added
	// to the request context first.
	span := sensor.Tracer().StartSpan("query-instana")
	defer span.Finish()

	ctx := instana.ContextWithSpan(context.Background(), span)
	req, err := http.NewRequest("GET", "https://www.instana.com", nil)
	if err != nil {
		log.Fatalf("failed to create a new request: %s", err)
	}

	_, err = client.Do(req.WithContext(ctx))
	if err != nil {
		log.Fatalf("failed to GET https://www.instana.com: %s", err)
	}
}
