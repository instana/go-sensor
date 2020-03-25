package instana

import (
	"context"
	"net/http"
	"runtime"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

type SpanSensitiveFunc func(span ot.Span)
type ContextSensitiveFunc func(span ot.Span, ctx context.Context)

// Sensor is used to inject tracing information into requests
type Sensor struct {
	tracer ot.Tracer
	logger LeveledLogger
}

// NewSensor creates a new instana.Sensor
func NewSensor(serviceName string) *Sensor {
	return NewSensorWithTracer(NewTracerWithOptions(
		&Options{
			Service: serviceName,
		},
	))
}

// NewSensorWithTracer returns a new instana.Sensor that uses provided tracer to report spans
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

// Executes the given ContextSensitiveFunc and executes it under the scope of a newly created context.Context,
// that provides access to the parent span as 'parentSpan'.
//
// Deprecated: please use instana.TracingHandlerFunc() to instrument an HTTP handler
func (s *Sensor) WithTracingContext(name string, w http.ResponseWriter, req *http.Request, f ContextSensitiveFunc) {
	s.WithTracingSpan(name, w, req, func(span ot.Span) {
		f(span, ContextWithSpan(req.Context(), span))
	})
}
