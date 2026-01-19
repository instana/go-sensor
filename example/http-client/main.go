// (c) Copyright IBM Corp. 2026

package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go/ext"
)

var server_url string
var collector instana.TracerLogger

func init() {
	collector = instana.InitCollector(&instana.Options{
		Service: "http-client",
		Tracer:  instana.DefaultTracerOptions(),
	})
}

func main() {
	flag.StringVar(&server_url, "s", "https://example.com", "Server address")
	flag.Parse()

	if server_url == "" {
		flag.Usage()
		os.Exit(2)
	}

	client := &http.Client{
		Transport: instana.RoundTripper(collector, nil),
	}

	req, err := http.NewRequest(http.MethodGet, server_url, nil)
	if err != nil {
		log.Fatalf("failed to create request: %s", err)
	}

	span := collector.Tracer().StartSpan("http-client-call")
	span.SetTag(string(ext.SpanKind), "entry")

	ctx := instana.ContextWithSpan(context.Background(), span)

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		log.Fatalf("failed to request to %v: %v", server_url, err)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	span.Finish()

	// Print the response status and body
	log.Println("Response Status: " + resp.Status)
	log.Println("Response Body: " + string(body))
	resp.Body.Close()

	// wait till user closes the program
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
}
