// (c) Copyright IBM Corp. 2023

package instaazurefunction_test

import (
	"net/http"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaazurefunction"
)

// This example demonstrates how to instrument a custom handler for Azure Functions
func Example_handler() {
	// Initialize a new collector.
	c := instana.InitCollector(&instana.Options{
		Service: "my-azf-sensor",
	})
	defer instana.ShutdownCollector()

	// Instrument your handler before passing it to the http router.
	http.HandleFunc("/api/azf-test", instaazurefunction.WrapFunctionHandler(c, handlerFn))
}

func handlerFn(w http.ResponseWriter, r *http.Request) {
	// ...
}
