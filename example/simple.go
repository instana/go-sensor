package main

import (
	"time"

	"github.com/instana/golang-sensor"
	ot "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/net/context"
)

const (
	Service = "golang-simple"
)

func simple(ctx context.Context) {
	parentSpan, ctx := ot.StartSpanFromContext(ctx, "parent")
	parentSpan.SetTag(string(ext.Component), Service)
	parentSpan.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCServerEnum))
	parentSpan.SetTag(string(ext.PeerHostname), "localhost")
	parentSpan.SetTag(string(ext.HTTPUrl), "/golang/simple/one")
	parentSpan.SetTag(string(ext.HTTPMethod), "GET")
	parentSpan.SetTag(string(ext.HTTPStatusCode), 200)
	parentSpan.LogFields(
		log.String("foo", "bar"))

	childSpan := ot.StartSpan("child", ot.ChildOf(parentSpan.Context()))
	childSpan.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCClientEnum))
	childSpan.SetTag(string(ext.PeerHostname), "localhost")
	childSpan.SetTag(string(ext.HTTPUrl), "/golang/simple/two")
	childSpan.SetTag(string(ext.HTTPMethod), "POST")
	childSpan.SetTag(string(ext.HTTPStatusCode), 204)
	childSpan.SetBaggageItem("someBaggage", "someValue")

	time.Sleep(450 * time.Millisecond)

	childSpan.Finish()

	time.Sleep(550 * time.Millisecond)

	parentSpan.Finish()
}

func main() {
	ot.InitGlobalTracer(instana.NewTracerWithOptions(&instana.Options{
		Service:  Service,
		LogLevel: instana.Debug}))

	go forever()
	go forever()
	select {}
}

func forever() {
	for {
		simple(context.Background())
	}
}
