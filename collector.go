// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
)

// TracerLogger represents the Instana Go collector and is composed by a tracer, a logger and a reference to the legacy sensor.
type TracerLogger interface {
	Tracer
	LeveledLogger
	LegacySensor() *Sensor
	Tracer() ot.Tracer
	Logger() LeveledLogger
}

// Collector is used to inject tracing information into requests
type Collector struct {
	t Tracer
	LeveledLogger
	*Sensor
}

var _ TracerLogger = (*Collector)(nil)

// InitCollector creates a new [Collector]
func InitCollector(opts *Options) TracerLogger {

	// if instana.C is already an instance of Collector, we just return
	if _, ok := C.(*Collector); ok {
		C.Warn("InitCollector was previously called. instana.C is reused")
		return C
	}

	if opts == nil {
		opts = &Options{
			Recorder: NewRecorder(),
		}
	}

	if opts.Recorder == nil {
		opts.Recorder = NewRecorder()
	}

	StartMetrics(opts)

	tracer := &tracerS{
		recorder: opts.Recorder,
	}

	C = &Collector{
		t:             tracer,
		LeveledLogger: defaultLogger,
		Sensor:        NewSensorWithTracer(tracer),
	}

	return C
}

// Extract() returns a SpanContext instance given `format` and `carrier`. It matches [opentracing.Tracer.Extract].
func (c *Collector) Extract(format interface{}, carrier interface{}) (ot.SpanContext, error) {
	return c.t.Extract(format, carrier)
}

// Inject() takes the `sm` SpanContext instance and injects it for
// propagation within `carrier`. The actual type of `carrier` depends on
// the value of `format`. It matches [opentracing.Tracer.Inject]
func (c *Collector) Inject(sm ot.SpanContext, format interface{}, carrier interface{}) error {
	return c.t.Inject(sm, format, carrier)
}

// Create, start, and return a new Span with the given `operationName` and
// incorporate the given StartSpanOption `opts`. (Note that `opts` borrows
// from the "functional options" pattern, per
// http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
//
// It matches [opentracing.Tracer.StartSpan].
func (c *Collector) StartSpan(operationName string, opts ...ot.StartSpanOption) ot.Span {
	return c.t.StartSpan(operationName, opts...)
}

// StartSpanWithOptions creates and starts a span by setting Instana relevant data within the span.
// It matches [instana.Tracer.StartSpanWithOptions].
func (c *Collector) StartSpanWithOptions(operationName string, opts ot.StartSpanOptions) ot.Span {
	return c.t.StartSpanWithOptions(operationName, opts)
}

// Options gets the current tracer options
// It matches [instana.Tracer.Options].
func (c *Collector) Options() TracerOptions {
	return c.t.Options()
}

// Flush sends all finished spans to the agent
// It matches [instana.Tracer.Flush].
func (c *Collector) Flush(ctx context.Context) error {
	return c.t.Flush(ctx)
}

// Debug logs a debug message by calling [LeveledLogger] underneath
func (c *Collector) Debug(v ...interface{}) {
	c.LeveledLogger.Debug(v...)
}

// Info logs an info message by calling [LeveledLogger] underneath
func (c *Collector) Info(v ...interface{}) {
	c.LeveledLogger.Info(v...)
}

// Warn logs a warning message by calling [LeveledLogger] underneath
func (c *Collector) Warn(v ...interface{}) {
	c.LeveledLogger.Warn(v...)
}

// Error logs a error message by calling [LeveledLogger] underneath
func (c *Collector) Error(v ...interface{}) {
	c.LeveledLogger.Error(v...)
}

// LegacySensor returns a reference to [Sensor] that can be used for old instrumentations that still require it.
//
// Example:
//
//	// Instrumenting HTTP incoming calls
//	c := instana.InitCollector("my-service")
//	http.HandleFunc("/", instana.TracingNamedHandlerFunc(c.LegacySensor(), "", "/{name}", handle))
func (c *Collector) LegacySensor() *Sensor {
	return c.Sensor
}

// Tracer returns an implementation of [opentracing.Tracer]
func (c *Collector) Tracer() ot.Tracer {
	return c.t
}

// Logger returns an implementation of [LeveledLogger]
func (c *Collector) Logger() LeveledLogger {
	return c.LeveledLogger
}
