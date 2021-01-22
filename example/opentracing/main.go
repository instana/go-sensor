// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var args struct {
	ListenAddr string
}

func main() {
	flag.StringVar(&args.ListenAddr, "l", os.Getenv("LISTEN_ADDR"), "Server listen address")
	flag.Parse()

	if args.ListenAddr == "" {
		flag.Usage()
		os.Exit(2)
	}

	// First we need to initialize an instance of Instana tracer and set it as the global tracer
	// for OpenTracing package
	tracer := instana.NewTracerWithOptions(instana.DefaultOptions())
	opentracing.InitGlobalTracer(tracer)

	http.HandleFunc("/", HandleRequest)

	log.Printf("example opentracing server is listening on %s...", args.ListenAddr)
	if err := http.ListenAndServe(args.ListenAddr, nil); err != nil {
		log.Fatalf("failed to start the server: %s", err)
	}
}

// HandleRequest responds with a "Hello, world!" message
func HandleRequest(w http.ResponseWriter, req *http.Request) {
	// Collect request parameters to add them to the entry HTTP span. We also need to make
	// sure that a proper span kind is set for the entry span, so that Instana could combine
	// it and its children into a call.
	opts := []opentracing.StartSpanOption{
		ext.SpanKindRPCServer,
		opentracing.Tags{
			"http.host":     req.Host,
			"http.method":   req.Method,
			"http.protocol": req.URL.Scheme,
			"http.path":     req.URL.Path,
		},
	}

	// Check if there is an ongoing trace context provided with request and use
	// it as a parent for our entry span to ensure continuation.
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
	if err != nil {
		opts = append(opts, ext.RPCServerOption(wireContext))
	}

	// Start the entry span adding collected tags and optional parent. The span name here
	// matters, as it allows Instana backend to classify the call as an HTTP one.
	span := opentracing.GlobalTracer().StartSpan("g.http", opts...)
	defer span.Finish()

	time.Sleep(300 * time.Millisecond)
	w.Write([]byte("Hello, world!\n"))
}
