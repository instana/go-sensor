package main

import (
	"time"

	"github.com/instana/golang-sensor"
	ot "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
)

const (
	SERVICE = "golang-rpc"
)

func rpc(ctx context.Context) {
	parentSpan, ctx := ot.StartSpanFromContext(ctx, "parentService.myCoolMethod")
	childSpan := ot.StartSpan("childService.anotherCoolMethod", ot.ChildOf(parentSpan.Context()))

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
		rpc(context.Background())
	}
}
