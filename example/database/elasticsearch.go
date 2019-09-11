package main

import (
	"math/rand"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"golang.org/x/net/context"
)

const (
	// Service - use a tracer level global service name
	Service = "Go-ES-Cluster-19m"
)

func fetchResults(parentSpan ot.Span) {
	// Making an ElasticSearch query
	childSpan := ot.StartSpan("elasticsearch", ot.ChildOf(parentSpan.Context()))

	// Set the apropriate tags for the span
	//
	// Client Span
	childSpan.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCClientEnum))
	// Specify the instance
	childSpan.SetTag(string(ext.DBInstance), "es-node-three")
	// Type
	childSpan.SetTag(string(ext.DBType), "elasticsearch")
	// The executed query
	childSpan.SetTag(string(ext.DBStatement), "{'query': {'match_all': {}}}")

	// Actual ElasticSearch calls (we'll just sleep for a random duration as an example)
	time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)

	childSpan.Finish()
}

func runJob(ctx context.Context) {
	parentSpan := ot.StartSpan("worker")
	parentSpan.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCServerEnum))

	// Fetch results from ElasticSearch
	fetchResults(parentSpan)

	parentSpan.Finish()
}

func main() {
	ot.InitGlobalTracer(instana.NewTracerWithOptions(&instana.Options{
		Service:  Service,
		LogLevel: instana.Info}))

	go doThisForever()
	select {}
}

func doThisForever() {
	for {
		runJob(context.Background())
		time.Sleep(2 * time.Second)
	}
}
