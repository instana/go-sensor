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

	client := &http.Client{
		Transport: instana.RoundTripper(col, nil),
	}

	// Instrument the trigger handler with Instana TracingHandlerFunc (entry span).
	// When you make outgoing HTTP client calls using r.Context(), the exit spans
	// will be attached as child spans of this entry span, avoiding the need for
	// INSTANA_ALLOW_ROOT_EXIT_SPAN=1.
	http.HandleFunc("/test", instana.TracingHandlerFunc(col, "/test", func(w http.ResponseWriter, r *http.Request) {
		statuses := []int{200, 401, 403, 404, 500}
		for _, status := range statuses {
			url := fmt.Sprintf("%s/respond?status=%d", upstream.URL, status)

			req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
			if err != nil {
				log.Printf("failed to create request: %v", err)
				continue
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("request error: %v", err)
				continue
			}
			resp.Body.Close()
			log.Printf("GET ?status=%d → HTTP %d", status, resp.StatusCode)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Triggered HTTP calls. Check the console and Instana dashboard for exit spans.")
	}))

	port := ":7070"
	log.Printf("Server running on http://localhost%s. Call http://localhost%s/test to trigger requests.", port, port)
	log.Fatal(http.ListenAndServe(port, nil))
}
