// (c) Copyright IBM Corp. 2026

// This example demonstrates the HTTP 4xx exit span error classification feature
// using the instafasthttp instrumentation package.
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
	"net"
	"os"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instafasthttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
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
		Service: "fasthttp-4xx-errors-example",
	})

	<-agentReady()

	// Start an in-process upstream server that responds with whatever status
	// code is requested via the "status" query parameter.
	ln := fasthttputil.NewInmemoryListener()
	upstreamServer := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			code := fasthttp.StatusOK
			fmt.Sscanf(string(ctx.QueryArgs().Peek("status")), "%d", &code)
			ctx.Response.SetStatusCode(code)
		},
	}
	go func() {
		if err := upstreamServer.Serve(ln); err != nil {
			log.Printf("upstream server error: %v", err)
		}
	}()
	defer ln.Close()

	// Prepare instrumented clients
	hc := &fasthttp.HostClient{
		Addr: "example.com",
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	}

	ic := instafasthttp.GetInstrumentedClient(col, &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	})

	// Server request handler with /test trigger endpoint
	handler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/test":
			uCtx := instafasthttp.UserContext(ctx)

			statuses := []int{200, 401, 403, 404, 500}

			// 1. HostClient + RoundTripper
			hc.Transport = instafasthttp.RoundTripper(uCtx, col, nil)
			for _, status := range statuses {
				req := fasthttp.AcquireRequest()
				resp := fasthttp.AcquireResponse()

				req.Header.SetMethod(fasthttp.MethodGet)
				req.SetRequestURIBytes([]byte(fmt.Sprintf("http://example.com/respond?status=%d", status)))

				if err := hc.Do(req, resp); err != nil {
					log.Printf("RoundTripper request error: %v", err)
				} else {
					log.Printf("RoundTripper: GET ?status=%d → HTTP %d", status, resp.StatusCode())
				}

				fasthttp.ReleaseRequest(req)
				fasthttp.ReleaseResponse(resp)
			}

			// 2. GetInstrumentedClient
			for _, status := range statuses {
				req := fasthttp.AcquireRequest()
				resp := fasthttp.AcquireResponse()

				req.Header.SetMethod(fasthttp.MethodGet)
				req.SetRequestURIBytes([]byte(fmt.Sprintf("http://example.com/respond?status=%d", status)))

				if err := ic.Do(uCtx, req, resp); err != nil {
					log.Printf("InstrumentedClient request error: %v", err)
				} else {
					log.Printf("InstrumentedClient: GET ?status=%d → HTTP %d", status, resp.StatusCode())
				}

				fasthttp.ReleaseRequest(req)
				fasthttp.ReleaseResponse(resp)
			}

			ctx.SetStatusCode(fasthttp.StatusOK)
			fmt.Fprintln(ctx, "Triggered fasthttp calls. Check the console and Instana dashboard for exit spans.")
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	port := ":7071"
	log.Printf("Server running on http://localhost%s. Call http://localhost%s/test to trigger requests.", port, port)
	log.Fatal(fasthttp.ListenAndServe(port, instafasthttp.TraceHandler(col, "fasthttp-server", "/test", handler)))
}
