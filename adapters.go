package instana

import (
	"context"
	"net/http"
	"runtime"

	"github.com/felixge/httpsnoop"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

type SpanSensitiveFunc func(span ot.Span)
type ContextSensitiveFunc func(span ot.Span, ctx context.Context)

type Sensor struct {
	tracer ot.Tracer
}

// Creates a new Instana sensor instance which can be used to
// inject tracing information into requests.
func NewSensor(serviceName string) *Sensor {
	return &Sensor{
		NewTracerWithOptions(
			&Options{
				Service: serviceName,
			},
		),
	}
}

// It is similar to TracingHandler in regards, that it wraps an existing http.HandlerFunc
// into a named instance to support capturing tracing information and data. It, however,
// provides a neater way to register the handler with existing frameworks by returning
// not only the wrapper, but also the URL-pattern to react on.
func (s *Sensor) TraceHandler(name, pattern string, handler http.HandlerFunc) (string, http.HandlerFunc) {
	return pattern, s.TracingHandler(name, handler)
}

// Wraps an existing http.HandlerFunc into a named instance to support capturing tracing
// information and response data.
func (s *Sensor) TracingHandler(name string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		s.WithTracingContext(name, w, req, func(span ot.Span, ctx context.Context) {
			// Capture response code for span
			hooks := httpsnoop.Hooks{
				WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
					return func(code int) {
						next(code)
						span.SetTag(string(ext.HTTPStatusCode), code)
					}
				},
			}

			// Add hooks to response writer
			wrappedWriter := httpsnoop.Wrap(w, hooks)

			// Serve original handler
			handler.ServeHTTP(wrappedWriter, req.WithContext(ctx))
		})
	}
}

// Wraps an existing http.Request instance into a named instance to inject tracing and span
// header information into the actual HTTP wire transfer.
func (s *Sensor) TracingHttpRequest(name string, parent, req *http.Request, client http.Client) (res *http.Response, err error) {
	var span ot.Span
	if parentSpan, ok := parent.Context().Value("parentSpan").(ot.Span); ok {
		span = s.tracer.StartSpan("client", ot.ChildOf(parentSpan.Context()))
	} else {
		span = s.tracer.StartSpan("client")
	}
	defer span.Finish()

	headersCarrier := ot.HTTPHeadersCarrier(req.Header)
	if err := s.tracer.Inject(span.Context(), ot.HTTPHeaders, headersCarrier); err != nil {
		return nil, err
	}

	res, err = client.Do(req.WithContext(context.Background()))

	span.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCClientEnum))
	span.SetTag(string(ext.PeerHostname), req.Host)
	span.SetTag(string(ext.HTTPUrl), req.URL.String())
	span.SetTag(string(ext.HTTPMethod), req.Method)
	span.SetTag(string(ext.HTTPStatusCode), res.StatusCode)

	if err != nil {
		span.LogFields(otlog.Error(err))
	}
	return
}

// Executes the given SpanSensitiveFunc and executes it under the scope of a child span, which is#
// injected as an argument when calling the function.
func (s *Sensor) WithTracingSpan(name string, w http.ResponseWriter, req *http.Request, f SpanSensitiveFunc) {
	wireContext, _ := s.tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))
	parentSpan := req.Context().Value("parentSpan")

	if name == "" {
		pc, _, _, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		name = f.Name()
	}

	var span ot.Span
	if ps, ok := parentSpan.(ot.Span); ok {
		span = s.tracer.StartSpan(
			name,
			ext.RPCServerOption(wireContext),
			ot.ChildOf(ps.Context()),
		)
	} else {
		span = s.tracer.StartSpan(
			name,
			ext.RPCServerOption(wireContext),
		)
	}

	span.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCServerEnum))
	span.SetTag(string(ext.PeerHostname), req.Host)
	span.SetTag(string(ext.HTTPUrl), req.URL.Path)
	span.SetTag(string(ext.HTTPMethod), req.Method)

	defer func() {
		// Capture outgoing headers
		s.tracer.Inject(span.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(w.Header()))

		// Make sure the span is sent in case we have to re-panic
		defer span.Finish()

		// Be sure to capture any kind of panic / error
		if err := recover(); err != nil {
			if e, ok := err.(error); ok {
				span.LogFields(otlog.Error(e))
			} else {
				span.LogFields(otlog.Object("error", err))
			}
			panic(err)
		}
	}()

	f(span)
}

// Executes the given ContextSensitiveFunc and executes it under the scope of a newly created context.Context,
// that provides access to the parent span as 'parentSpan'.
func (s *Sensor) WithTracingContext(name string, w http.ResponseWriter, req *http.Request, f ContextSensitiveFunc) {
	s.WithTracingSpan(name, w, req, func(span ot.Span) {
		ctx := context.WithValue(req.Context(), "parentSpan", span)
		f(span, ctx)
	})
}
