// (c) Copyright IBM Corp. 2024

package instafasthttp_test

import (
	"context"
	"fmt"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instafasthttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }

func BenchmarkTracingHandlerFunc(b *testing.B) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	s := instana.NewSensorWithTracer(tracer)
	// defer instana.ShutdownSensor()

	h := instafasthttp.TraceHandler(s, "/action", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		fmt.Fprintf(ctx, "Ok")
	})

	server := &fasthttp.Server{
		Handler: h,
	}

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(ln); err != nil {
			b.Errorf("unexpected error: %v", err)
		}
	}()

	// req := httptest.NewRequest(http.MethodGet, "/test?q=term", nil)

	b.ResetTimer()

	for i := 0; i < 10; i++ {
		conn, err := ln.Dial()
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}

		if _, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: google.com\r\n\r\n")); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
