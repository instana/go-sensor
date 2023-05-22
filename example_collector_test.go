// (c) Copyright IBM Corp. 2023

package instana_test

import (
	"fmt"
	"net/http"
	"time"

	instana "github.com/instana/go-sensor"
)

func Example_collectorBasicUsage() {
	// Initialize the collector
	c := instana.InitCollector(&instana.Options{
		// If Service is not provided, the executable filename will be used instead.
		// We recommend that Service be set.
		Service: "my-go-app",
	})

	// Instrument something
	sp := c.StartSpan("my_span")

	time.Sleep(time.Second * 3)

	sp.Finish()
}

func Example_collectorWithHTTPServer() {
	c := instana.InitCollector(&instana.Options{
		Service: "my-go-app",
	})

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Ok")
	}

	http.HandleFunc("/foo", instana.TracingHandlerFunc(c.LegacySensor(), "/foo", handler))
}
