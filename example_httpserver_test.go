// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"log"
	"net/http"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
)

// This example shows how to instrument an HTTP server with Instana tracing
func Example_tracingNamedHandlerFunc() {
	c := instana.InitCollector(&instana.Options{
		Service: "my-http-server",
	})
	defer instana.ShutdownCollector()

	// To instrument a handler function, pass it as an argument to instana.TracingHandlerFunc()
	http.HandleFunc("/", instana.TracingHandlerFunc(c, "/", func(w http.ResponseWriter, req *http.Request) {
		// Extract the parent span and use its tracer to initialize any child spans to trace the calls
		// inside the handler, e.g. database queries, 3rd-party API requests, etc.
		if parent, ok := instana.SpanFromContext(req.Context()); ok {
			sp := parent.Tracer().StartSpan("index", opentracing.ChildOf(parent.Context()))
			defer sp.Finish()
		}

		// ...

		w.Write([]byte("OK"))
	}))

	// In case your handler is implemented as an http.Handler, pass its ServeHTTP method instead.
	// You can also use instana.TracingNamedHandlerFunc() to provide a unique route identifier to
	// group the calls to this route later in Instana UI.
	http.HandleFunc("/files", instana.TracingNamedHandlerFunc(c, "index", "/:path", http.FileServer(http.Dir("./")).ServeHTTP))

	if err := http.ListenAndServe(":0", nil); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}

// This example demonstrates how to instrument a 3rd-party HTTP router that uses pattern matching to make sure
// the original path template is forwarded to Instana
func Example_httpRoutePatternMatching() {
	c := instana.InitCollector(&instana.Options{
		Service: "my-http-server",
	})
	defer instana.ShutdownCollector()

	// Initialize your router. Here for simplicity we use stdlib http.ServeMux, however you
	// can use any router that supports http.Handler/http.HandlerFunc, such as github.com/gorilla/mux
	r := http.NewServeMux()

	// Wrap the handler using instana.TracingHandlerFunc() and pass the path template
	// used to route requests to it. This value will be attached to the span and displayed
	// in UI as `http.path_tpl` if the request path is different (we assume that this request
	// has been routed to the handler via a matched pattern)
	r.HandleFunc("/articles/{category}/{id:[0-9]+}", instana.TracingHandlerFunc(
		c,
		"/articles/{category}/{id:[0-9]+}",
		func(w http.ResponseWriter, req *http.Request) {
			// ...
		},
	))

	if err := http.ListenAndServe(":0", nil); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}
