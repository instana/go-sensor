package main

import (
	"math/rand"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/net/context"
)

const (
	// Service - use a tracer level global service name
	Service = "Go-Overlord"
)

func simple(ctx context.Context) {
	// Handling an incoming request
	parentSpan := ot.StartSpan("asteroid")
	parentSpan.SetTag(string(ext.Component), "Go simple example app")
	parentSpan.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCServerEnum))
	parentSpan.SetTag(string(ext.HTTPUrl), "https://asteroid.svc.io/golang/api/v2")
	parentSpan.SetTag(string(ext.HTTPMethod), "GET")
	parentSpan.SetTag(string(ext.HTTPStatusCode), uint16(200))
	parentSpan.LogFields(log.String("foo", "bar"))

	// Making an HTTP request
	childSpan := ot.StartSpan("spacedust", ot.ChildOf(parentSpan.Context()))
	childSpan.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCClientEnum))
	childSpan.SetTag(string(ext.HTTPUrl), "https://meteor.svc.io/golang/api/v2")
	childSpan.SetTag(string(ext.HTTPMethod), "POST")
	childSpan.SetTag(string(ext.HTTPStatusCode), 204)
	childSpan.SetBaggageItem("someBaggage", "someValue")

	// Make the HTTP request (we'll just sleep for a random duration as an example)
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	childSpan.Finish()

	// Rendering a template with custom tags
	renderSpan := ot.StartSpan("render", ot.ChildOf(childSpan.Context()))
	renderSpan.SetTag("type", "layout")
	renderSpan.SetTag("name", "application_layout.erb")

	// Render the template (we'll just sleep for a random duration as an example)
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	renderSpan.Finish()

	// Additional work (we'll just sleep for a random duration as an example)
	time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)
	parentSpan.Finish()
}

func main() {
	ot.InitGlobalTracer(instana.NewTracerWithOptions(&instana.Options{
		Service:  Service,
		LogLevel: instana.Info}))

	go forever()
	select {}
}

func forever() {
	for {
		simple(context.Background())
		time.Sleep(2 * time.Second)
	}
}
