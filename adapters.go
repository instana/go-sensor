// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2019

package instana

import (
	"context"
	"net/http"
	"runtime"

	"github.com/opentracing/opentracing-go"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var _ TracerLogger = (*Sensor)(nil)

type SensorLogger interface {
	Tracer() ot.Tracer
	Logger() LeveledLogger
	SetLogger(l LeveledLogger)
}

// SpanSensitiveFunc is a function executed within a span context
//
// Deprecated: use instana.ContextWithSpan() and instana.SpanFromContext() to inject and retrieve spans
type SpanSensitiveFunc func(span ot.Span)

// ContextSensitiveFunc is a SpanSensitiveFunc that also takes context.Context
//
// Deprecated: use instana.ContextWithSpan() and instana.SpanFromContext() to inject and retrieve spans
type ContextSensitiveFunc func(span ot.Span, ctx context.Context)

// Tracer extends the opentracing.Tracer interface
type Tracer interface {
	opentracing.Tracer

	// Options gets the current tracer options
	Options() TracerOptions
	// Flush sends all finished spans to the agent
	Flush(context.Context) error
	// StartSpanWithOptions starts a span with the given options and return the span reference
	StartSpanWithOptions(string, ot.StartSpanOptions) ot.Span
}

// Sensor is used to inject tracing information into requests
type Sensor struct {
	tracer ot.Tracer
	logger LeveledLogger
}

// NewSensor creates a new [Sensor]
func NewSensor(serviceName string) *Sensor {
	return NewSensorWithTracer(NewTracerWithOptions(
		&Options{
			Service: serviceName,
		},
	))
}

// NewSensorWithTracer returns a new [Sensor] that uses provided tracer to report spans
func NewSensorWithTracer(tracer ot.Tracer) *Sensor {
	return &Sensor{
		tracer: tracer,
		logger: defaultLogger,
	}
}

// Tracer returns the tracer instance for this sensor
func (s *Sensor) Tracer() ot.Tracer {
	return s.tracer
}

// Logger returns the logger instance for this sensor
func (s *Sensor) Logger() LeveledLogger {
	return s.logger
}

// SetLogger sets the logger for this sensor
func (s *Sensor) SetLogger(l LeveledLogger) {
	s.logger = l
}

// TraceHandler is similar to TracingHandler in regards, that it wraps an existing http.HandlerFunc
// into a named instance to support capturing tracing information and data. The returned values are
// compatible with handler registration methods, e.g. http.Handle()
//
// Deprecated: please use instana.TracingHandlerFunc() instead
func (s *Sensor) TraceHandler(name, pattern string, handler http.HandlerFunc) (string, http.HandlerFunc) {
	return pattern, s.TracingHandler(name, handler)
}

// TracingHandler wraps an existing http.HandlerFunc into a named instance to support capturing tracing
// information and response data
//
// Deprecated: please use instana.TracingHandlerFunc() instead
func (s *Sensor) TracingHandler(name string, handler http.HandlerFunc) http.HandlerFunc {
	return TracingHandlerFunc(s, name, handler)
}

// TracingHttpRequest wraps an existing http.Request instance into a named instance to inject tracing and span
// header information into the actual HTTP wire transfer
//
// Deprecated: please use instana.RoundTripper() instead
func (s *Sensor) TracingHttpRequest(name string, parent, req *http.Request, client http.Client) (*http.Response, error) {
	client.Transport = RoundTripper(s, client.Transport)
	return client.Do(req.WithContext(context.Background()))
}

// WithTracingSpan takes the given SpanSensitiveFunc and executes it under the scope of a child span, which is
// injected as an argument when calling the function. It uses the name of the caller as a span operation name
// unless a non-empty value is provided
//
// Deprecated: please use instana.TracingHandlerFunc() to instrument an HTTP handler
func (s *Sensor) WithTracingSpan(operationName string, w http.ResponseWriter, req *http.Request, f SpanSensitiveFunc) {
	if operationName == "" {
		pc, _, _, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		operationName = f.Name()
	}

	opts := []ot.StartSpanOption{
		ext.SpanKindRPCServer,

		ot.Tags{
			string(ext.PeerHostname): req.Host,
			string(ext.HTTPUrl):      req.URL.Path,
			string(ext.HTTPMethod):   req.Method,
		},
	}

	wireContext, err := s.tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))
	switch err {
	case nil:
		opts = append(opts, ext.RPCServerOption(wireContext))
	case ot.ErrSpanContextNotFound:
		s.Logger().Debug("no span context provided with ", req.Method, " ", req.URL.Path)
	case ot.ErrUnsupportedFormat:
		s.Logger().Info("unsupported span context format provided with ", req.Method, " ", req.URL.Path)
	default:
		s.Logger().Warn("failed to extract span context from the request:", err)
	}

	if ps, ok := SpanFromContext(req.Context()); ok {
		opts = append(opts, ot.ChildOf(ps.Context()))
	}

	span := s.tracer.StartSpan(operationName, opts...)
	defer span.Finish()

	defer func() {
		// Capture outgoing headers
		s.tracer.Inject(span.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(w.Header()))

		// Be sure to capture any kind of panic / error
		if err := recover(); err != nil {
			if e, ok := err.(error); ok {
				span.LogFields(otlog.Error(e))
			} else {
				span.LogFields(otlog.Object("error", err))
			}

			// re-throw the panic
			panic(err)
		}
	}()

	f(span)
}

// WithTracingContext executes the given ContextSensitiveFunc and executes it under the scope of a newly created context.Context,
// that provides access to the parent span as 'parentSpan'.
//
// Deprecated: please use instana.TracingHandlerFunc() to instrument an HTTP handler
func (s *Sensor) WithTracingContext(name string, w http.ResponseWriter, req *http.Request, f ContextSensitiveFunc) {
	s.WithTracingSpan(name, w, req, func(span ot.Span) {
		f(span, ContextWithSpan(req.Context(), span))
	})
}

// Compliance with TracerLogger

// Extract() returns a SpanContext instance given `format` and `carrier`. It matches [opentracing.Tracer.Extract].
func (s *Sensor) Extract(format interface{}, carrier interface{}) (ot.SpanContext, error) {
	return s.tracer.Extract(format, carrier)
}

// Inject() takes the `sm` SpanContext instance and injects it for
// propagation within `carrier`. The actual type of `carrier` depends on
// the value of `format`. It matches [opentracing.Tracer.Inject]
func (s *Sensor) Inject(sm ot.SpanContext, format interface{}, carrier interface{}) error {
	return s.tracer.Inject(sm, format, carrier)
}

// Create, start, and return a new Span with the given `operationName` and
// incorporate the given StartSpanOption `opts`. (Note that `opts` borrows
// from the "functional options" pattern, per
// http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
//
// It matches [opentracing.Tracer.StartSpan].
func (s *Sensor) StartSpan(operationName string, opts ...ot.StartSpanOption) ot.Span {
	return s.tracer.StartSpan(operationName, opts...)
}

// StartSpanWithOptions creates and starts a span by setting Instana relevant data within the span.
// It matches [instana.Tracer.StartSpanWithOptions].
func (s *Sensor) StartSpanWithOptions(operationName string, opts ot.StartSpanOptions) ot.Span {
	if t, ok := s.tracer.(Tracer); ok {
		return t.StartSpanWithOptions(operationName, opts)
	}

	s.logger.Warn("Sensor.StartSpanWithOptions() not implemented by interface: ", s.tracer, " - returning nil")

	return nil
}

// Options gets the current tracer options
// It matches [instana.Tracer.Options].
func (s *Sensor) Options() TracerOptions {
	if t, ok := s.tracer.(Tracer); ok {
		return t.Options()
	}

	s.logger.Warn("Sensor.Options() not implemented by interface: ", s.tracer, " - returning DefaultTracerOptions()")

	return DefaultTracerOptions()
}

// Flush sends all finished spans to the agent
// It matches [instana.Tracer.Flush].
func (s *Sensor) Flush(ctx context.Context) error {
	if t, ok := s.tracer.(Tracer); ok {
		return t.Flush(ctx)
	}

	s.logger.Warn("Sensor.Flush() not implemented by interface: ", s.tracer, " - returning nil")

	return nil
}

// Debug logs a debug message by calling [LeveledLogger] underneath
func (s *Sensor) Debug(v ...interface{}) {
	s.logger.Debug(v...)
}

// Info logs an info message by calling [LeveledLogger] underneath
func (s *Sensor) Info(v ...interface{}) {
	s.logger.Info(v...)
}

// Warn logs a warning message by calling [LeveledLogger] underneath
func (s *Sensor) Warn(v ...interface{}) {
	s.logger.Warn(v...)
}

// Error logs a error message by calling [LeveledLogger] underneath
func (s *Sensor) Error(v ...interface{}) {
	s.logger.Error(v...)
}

// LegacySensor returns a reference to [Sensor].
func (s *Sensor) LegacySensor() *Sensor {
	return s
}
