// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"log"
	"net/http"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go/ext"
)

// This example shows how to instrument an HTTP client with Instana tracing
func Example_roundTripper() {
	sensor := instana.NewSensor("my-http-client")

	// Wrap the original http.Client transport with instana.RoundTripper().
	// The http.DefaultTransport will be used if there was no transport provided.
	client := &http.Client{
		Transport: instana.RoundTripper(sensor, nil),
	}

	// The call should always start with an entry span (https://docs.instana.io/quick_start/custom_tracing/#always-start-new-traces-with-entry-spans)
	// Normally this would be your HTTP/GRPC/message queue request span, but here we need to
	// create it explicitly.
	sp := sensor.Tracer().StartSpan("client-call")
	sp.SetTag(string(ext.SpanKind), "entry")

	req, err := http.NewRequest(http.MethodGet, "https://www.instana.com", nil)
	if err != nil {
		log.Fatalf("failed to create request: %s", err)
	}

	// Inject the parent span into request context
	ctx := instana.ContextWithSpan(context.Background(), sp)

	// Use your instrumented http.Client to propagate tracing context with the request
	_, err = client.Do(req.WithContext(ctx))
	if err != nil {
		log.Fatalf("failed to GET https://www.instana.com: %s", err)
	}
}
