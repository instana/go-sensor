package instana_test

import (
	"context"
	"net/http"

	instana "github.com/instana/go-sensor"
)

// This example demonstrates how to instrument an HTTP handler with Instana and register it
// in http.DefaultServeMux
func ExampleTracingHandlerFunc() {
	// Here we initialize a new instance of instana.Sensor, however it is STRONGLY recommended
	// to use a single instance throughout your application
	sensor := instana.NewSensor("my-http-server")

	http.HandleFunc("/", instana.TracingHandlerFunc(sensor, "root", func(w http.ResponseWriter, req *http.Request) {
		// handler code
	}))
}

// This example demonstrates how to instrument an HTTP client with Instana
func ExampleRoundTripper() {
	// Here we initialize a new instance of instana.Sensor, however it is STRONGLY recommended
	// to use a single instance throughout your application
	sensor := instana.NewSensor("my-http-client")
	span := sensor.Tracer().StartSpan("entry")

	// http.DefaultTransport is used as a default RoundTripper, however you can provide
	// your own implementation
	client := &http.Client{
		Transport: instana.RoundTripper(sensor, nil),
	}

	// Inject parent span into therequest context
	ctx := instana.ContextWithSpan(context.Background(), span)
	req, _ := http.NewRequest("GET", "https://www.instana.com", nil)

	// Execute request as usual
	client.Do(req.WithContext(ctx))
}
