// (c) Copyright IBM Corp. 2023

// Package instaazurefunction provides Instana tracing instrumentation for
// Microsoft Azure Functions
package instaazurefunction

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

const (
	customRuntime       string = "custom"
	azfFlushMaxRetries         = 5
	azfFlushRetryPeriod        = 50 * time.Millisecond
	azfTimeoutThreshold        = 2 * 100 * time.Millisecond
)

// WrapFunctionHandler wraps the http handler and add instrumentation data for the specified handlers
func WrapFunctionHandler(sensor *instana.Sensor, handler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		tracer := sensor.Tracer()
		opts := []ot.StartSpanOption{
			ot.Tags{
				"azf.runtime": customRuntime,
			},
		}

		ctx := req.Context()

		var span ot.Span
		if parent, ok := instana.SpanFromContext(ctx); ok {
			opts = append(opts, ot.ChildOf(parent.Context()))
		}
		span = tracer.StartSpan("azf", opts...)

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

		body, err := copyRequestBody(req)
		if err != nil {
			span.SetTag("azf.error", err.Error())
			span.LogFields(otlog.Object("error", err.Error()))
		}

		spanData, err := extractSpanData(body)
		if err != nil {
			span.SetTag("azf.error", err.Error())
			span.LogFields(otlog.Object("error", err.Error()))
		}

		span.SetTag("azf.functionname", spanData.FunctionName)
		span.SetTag("azf.triggername", spanData.TriggerType)

		// Here we create a separate context.Context to finalize and send the span. This context
		// supposed to be canceled once the wrapped handler is done.
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
	sensor.Logger().Debug("flushing trace data to the endpoint")

	if tr, ok := sensor.Tracer().(instana.Tracer); ok {
		var i int
		for {
			if err := tr.Flush(context.Background()); err != nil {
				if err == instana.ErrAgentNotReady && i < maxRetries {
					i++
					time.Sleep(retryPeriod)
					continue
				}

				sensor.Logger().Error("failed to send traces to the endpoint. Error:", err)
			}

			sensor.Logger().Debug("data sent to the endpoint")
			break
		}
	}
}

// copyRequestBody clones the request body and returns it
func copyRequestBody(req *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(req.Body)

	//request body will be empty if we do not assign it back
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return body, err
}
