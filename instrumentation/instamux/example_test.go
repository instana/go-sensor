// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

//go:build go1.12
// +build go1.12

package instamux_test

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instamux"
)

// This example shows how to instrument an HTTP server that uses github.com/gorilla/mux with Instana
func Example() {
	sensor := instana.NewSensor("my-web-server")
	r := mux.NewRouter()

	// Add an instrumentation middleware to the router. This middleware will be applied to all handlers
	// registered with this instance.
	instamux.AddMiddleware(sensor, r)

	// Use mux.Router to register request handlers as usual
	r.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
		// ...
	})

	if err := http.ListenAndServe(":0", nil); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}
