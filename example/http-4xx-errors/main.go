// (c) Copyright IBM Corp. 2026

// This example demonstrates the HTTP 4xx exit span error classification feature.
//
// By default, Instana does not treat 4xx responses as errors on exit spans.
// This program opts in via a local config file (config.yaml) so that only
// 401 and 403 responses are marked as errors on exit spans.
//
// # Other ways to configure the same behaviour
//
// A) Flat environment variables (no config file needed):
//
//	INSTANA_TRACING_HTTP_EXIT_CLASSIFY_AS_ERRORS=401,403
//	# or to mark ALL 4xx as errors:
//	INSTANA_TRACING_HTTP_EXIT_CLASSIFY_ALL_4XX_AS_ERRORS=true
//
// B) In-code options (overrides agent config, loses to env vars):
//
//	instana.Options{
//	    Tracer: instana.TracerOptions{
//	        HTTP: struct{ Exit instana.HTTPExitSettings }{
//	            Exit: instana.HTTPExitSettings{
//	                ClassifyAsErrors: []int{401, 403},
//	            },
//	        },
//	    },
//	}
//
// C) Agent configuration.yaml (lowest priority, applied at announce time):
//
//	com.instana.tracing:
//	  http:
//	    exit:
//	      classify-as-errors:
//	        - 401
//	        - 403
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	instana "github.com/instana/go-sensor"
)

func agentReady() chan bool {
	ch := make(chan bool)

	go func() {
		for {
			if instana.Ready() {
				ch <- true
			}
		}
	}()

	return ch
}

func main() {
	// Point the tracer at the local config file (config.yaml).
	// The file configures: tracing.http.exit.classify-as-errors: [401, 403]
	os.Setenv("INSTANA_CONFIG_PATH", "config.yaml")

	col := instana.InitCollector(&instana.Options{
		Service: "http-4xx-errors-example",
	})

	<-agentReady()

	// Start a small upstream server that responds with whatever status
	// code is requested via the "status" query parameter.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusOK
		fmt.Sscanf(r.URL.Query().Get("status"), "%d", &code)
		w.WriteHeader(code)
	}))
	defer upstream.Close()

	// Wrap the HTTP client with the Instana RoundTripper.
	// Every outgoing request will be recorded as an exit span.
	// When a 4xx response matches the classify-as-errors list,
	// span.ec is set to 1 and span.data.http.error is populated.
	client := &http.Client{
		Transport: instana.RoundTripper(col, nil),
	}

	statuses := []int{200, 401, 403, 404, 500}
	for _, status := range statuses {
		url := fmt.Sprintf("%s/respond?status=%d", upstream.URL, status)
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("request error: %v", err)
			continue
		}
		resp.Body.Close()
		log.Printf("GET ?status=%d → HTTP %d", status, resp.StatusCode)
	}

	// Expected span.ec with config.yaml (classify-as-errors: [401, 403]):
	//
	//   status 200 → ec=0  (success, not an error)
	//   status 401 → ec=1  (in the classify-as-errors list → error)
	//   status 403 → ec=1  (in the classify-as-errors list → error)
	//   status 404 → ec=0  (4xx but NOT in the list → not an error)
	//   status 500 → ec=1  (5xx is always an error regardless of 4xx config)

	time.Sleep(time.Minute * 10)

	log.Println("Done. Check your Instana dashboard for the exit spans.")
}
