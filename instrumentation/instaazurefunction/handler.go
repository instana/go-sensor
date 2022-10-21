// (c) Copyright IBM Corp. 2022
// (c) Copyright Instana Inc. 2022

// Package instaazurefunction provides Instana tracing instrumentation for
// Microsoft Azure Functions
package instaazurefunction

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

const (
	runtime             string = "custom"
	azfFlushMaxRetries         = 5
	azfFlushRetryPeriod        = 50 * time.Millisecond
	azfTimeoutThreshold        = 100 * time.Millisecond
)

// WrapFunctionHandler wraps the http handler and add instrumentation data for the specified handlers
func WrapFunctionHandler(sensor *instana.Sensor, handler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		tracer := sensor.Tracer()
		opts := []ot.StartSpanOption{
			ot.Tags{
				"azf.runtime": runtime,
			},
		}

		span := tracer.StartSpan("azf", opts...)

		defer func() {
			// Be sure to capture any kind of panic/error
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					span.SetTag("azf.error", e.Error())
					span.LogFields(otlog.Error(e))
				} else {
					span.SetTag("azf.error", err)
					span.LogFields(otlog.Object("error", err))
				}

				// re-throw the panic
				panic(err)
			}
		}()

		//clone the request body
		body, err := io.ReadAll(req.Body)

		if err != nil {
			span.SetTag("azf.error", err)
			span.LogFields(otlog.Object("error", err))
		}

		//request body will be empty if we do not assign it back
		req.Body = io.NopCloser(bytes.NewBuffer(body))

		trigger, method := extractSpanData(body)

		if trigger == unknownType || method == "" {
			span.SetTag("azf.error", "invalid type/method")
			span.LogFields(otlog.Object("error", err))
		}
		span.SetTag("azf.methodname", method)
		span.SetTag("azf.trigger", trigger)

		// Here we create a separate context.Context to finalize and send the span. This context
		// supposed to be canceled once the wrapped handler is done.
		ctx := req.Context()
		traceCtx, cancelTraceCtx := context.WithCancel(ctx)

		// In case there is a deadline defined for this invocation, we should finish the span just before
		// the function times out to send the span data.
		originalDeadline, deadlineDefined := ctx.Deadline()
		if deadlineDefined {
			traceCtx, cancelTraceCtx = context.WithDeadline(ctx, originalDeadline.Add(-azfTimeoutThreshold))
		}

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-traceCtx.Done()

			if traceCtx.Err() == context.DeadlineExceeded {
				remainingTime := time.Until(originalDeadline).Truncate(time.Millisecond)
				sensor.Logger().Debug("heuristic timeout detection was triggered with ", remainingTime, " left")
			}

			span.Finish()
			flushAgent(sensor, azfFlushRetryPeriod, azfFlushMaxRetries)
		}()

		handler(w, req.WithContext(instana.ContextWithSpan(ctx, span)))

		cancelTraceCtx()
		wg.Wait()
	}
}

// flushAgent sends the instrumentation data to the serverless endpoint
func flushAgent(sensor *instana.Sensor, retryPeriod time.Duration, maxRetries int) {
	sensor.Logger().Debug("flushing trace data")

	if tr, ok := sensor.Tracer().(instana.Tracer); ok {
		var i int
		for {
			if err := tr.Flush(context.Background()); err != nil {
				if err == instana.ErrAgentNotReady && i < maxRetries {
					i++
					time.Sleep(retryPeriod)
					continue
				}

				sensor.Logger().Error("failed to send traces:", err)
			}

			sensor.Logger().Debug("data sent")
			break
		}
	}
}
