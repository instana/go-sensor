package main

import (
	"time"

	"github.com/instana/golang-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/net/context"
)

const (
	SERVICE = "golang-simple"
)

func simple(ctx context.Context) {
	parentSpan, ctx := ot.StartSpanFromContext(ctx, "parent")
	parentSpan.LogFields(
		log.String("type", instana.HTTP_SERVER),
		log.Object("data", &instana.Data{
			Http: &instana.HttpData{
				Host:   "localhost",
				Url:    "/golang/simple/one",
				Status: 200,
				Method: "GET"}}))

	childSpan := ot.StartSpan("child", ot.ChildOf(parentSpan.Context()))
	childSpan.SetTag("component", "bar")
	childSpan.LogFields(
		log.String("type", instana.HTTP_CLIENT),
		log.Object("data", &instana.Data{
			Http: &instana.HttpData{
				Host:   "localhost",
				Url:    "/golang/simple/two",
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
		Service:  SERVICE,
		LogLevel: instana.DEBUG}))

	go forever()
	select {}
}

func forever() {
	for {
		simple(context.Background())
	}
}
