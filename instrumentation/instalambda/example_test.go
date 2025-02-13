// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instalambda_test

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instalambda"
)

// This example demonstrates how to instrument a handler function with Instana
func Example() {
	// Initialize a new collector
	c := instana.InitCollector(&instana.Options{
		Service: "my-go-lambda",
	})
	defer instana.ShutdownCollector()

	// Create a new instrumented lambda.Handler from your handle function
	h := instalambda.NewHandler(func(ctx context.Context) (string, error) {
		// If your handler function takes context.Context as a first argument,
		// instrumentation will inject the parent span into it, so you can continue
		// the trace beyond your Lambda handler, e.g. when making external HTTP calls,
		// database queries, etc.
		if parent, ok := instana.SpanFromContext(ctx); ok {
			sp := parent.Tracer().StartSpan("internal")
			defer sp.Finish()
		}

		return "Hello, ƛ!", nil
	}, c)

	// Pass the instrumented handler to lambda.StartHandler()
	lambda.StartHandler(h)
}

// This example demonstrates how to instrument a handler function invoked with an API Gateway event
func Example_apiGateway() {
	// Initialize a new collector
	c := instana.InitCollector(&instana.Options{
		Service: "my-go-lambda",
	})
	defer instana.ShutdownCollector()

	// Create a new instrumented lambda.Handler from your handle function
	h := instalambda.NewHandler(func(event *events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       "Hello, ƛ!",
		}, nil
	}, c)

	// Pass the instrumented handler to lambda.StartHandler()
	lambda.StartHandler(h)
}

// To instrument a handler function, create a new lambda.Handler using instalambda.NewHandler() and pass it to
// lambda.StartHandler()
func ExampleNewHandler() {
	// Initialize a new collector
	c := instana.InitCollector(&instana.Options{
		Service: "my-go-lambda",
	})
	defer instana.ShutdownCollector()

	// Create a new instrumented lambda.Handler from your handle function
	h := instalambda.NewHandler(func() (string, error) {
		return "Hello, ƛ!", nil
	}, c)

	// Pass the instrumented handler to lambda.StartHandler()
	lambda.StartHandler(h)
}

// To instrument a lambda.Handler, instrument it using instalambda.WrapHandler before passing to
// lambda.StartHandler()
func ExampleWrapHandler() {
	// Initialize a new collector
	c := instana.InitCollector(&instana.Options{
		Service: "my-go-lambda",
	})
	defer instana.ShutdownCollector()

	h := Handler{
		// ...
	}

	// Instrument your handler before passing it to lambda.StartHandler()
	lambda.StartHandler(instalambda.WrapHandler(h, c))
}

// Handler is an example AWS Lambda handler
type Handler struct{}

// Invoke handles AWS Lambda events
func (Handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	return []byte("Hello, ƛ!"), nil
}
