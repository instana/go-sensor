package main

import (
	"time"

	"github.com/instana/golang-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/net/context"
)

const (
	Service = "golang-simple"
)

func simple(ctx context.Context) {
	parentSpan, ctx := ot.StartSpanFromContext(ctx, "parent")
	parentSpan.LogFields(
		log.String("type", instana.HTTPServer),
		log.Object("data", &instana.Data{
			HTTP: &instana.HTTPData{
				Host:   "localhost",
				URL:    "/golang/simple/one",
				Status: 200,
				Method: "GET"}}))

	childSpan := ot.StartSpan("child", ot.ChildOf(parentSpan.Context()))
	childSpan.LogFields(
		log.String("type", instana.HTTPClient),
		log.Object("data", &instana.Data{
			HTTP: &instana.HTTPData{
				Host:   "localhost",
				URL:    "/golang/simple/two",
				Status: 204,
				Method: "POST"}}))
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
	select {}
}

func forever() {
	for {
		simple(context.Background())
	}
}
