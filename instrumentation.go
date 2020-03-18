package instana

import (
	"net/http"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// TracingHandlerFunc is an HTTP middleware that captures the tracing data and ensures
// trace context propagation via OpenTracing headers
func TracingHandlerFunc(sensor *Sensor, name string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		opts := []ot.StartSpanOption{
			ext.SpanKindRPCServer,
			ot.Tags{
				string(ext.PeerHostname): req.Host,
				string(ext.HTTPUrl):      req.URL.Path,
				string(ext.HTTPMethod):   req.Method,
			},
		}

		tracer := sensor.Tracer()
		if ps, ok := SpanFromContext(req.Context()); ok {
			tracer = ps.Tracer()
			opts = append(opts, ot.ChildOf(ps.Context()))
		}

		wireContext, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))
		switch err {
		case nil:
			opts = append(opts, ext.RPCServerOption(wireContext))
		case ot.ErrSpanContextNotFound:
			sensor.Logger().Debug("no span context provided with", req.Method, req.URL.Path)
		case ot.ErrUnsupportedFormat:
			sensor.Logger().Info("unsupported span context format provided with", req.Method, req.URL.Path)
		default:
			sensor.Logger().Warn("failed to extract span context from the request:", err)
		}

		span := tracer.StartSpan(name, opts...)
		defer span.Finish()

		defer func() {
			// Be sure to capture any kind of panic / error
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					span.SetTag("message", e.Error())
					span.SetTag("http.error", e.Error())
					span.LogFields(otlog.Error(e))
				} else {
					span.SetTag("message", err)
					span.SetTag("http.error", err)
					span.LogFields(otlog.Object("error", err))
				}

				span.SetTag(string(ext.HTTPStatusCode), http.StatusInternalServerError)

				// re-throw the panic
				panic(err)
			}
		}()

		wrapped := &statusCodeRecorder{ResponseWriter: w}
		tracer.Inject(span.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(wrapped.Header()))

		ctx = ContextWithSpan(ctx, span)
		handler(wrapped, req.WithContext(ctx))

		if wrapped.Status > 0 {
			span.SetTag(string(ext.HTTPStatusCode), wrapped.Status)
		}
	}
}

// wrapper over http.ResponseWriter to spy the returned status code
type statusCodeRecorder struct {
	http.ResponseWriter
	Status int
}

func (rec *statusCodeRecorder) WriteHeader(status int) {
	rec.Status = status
	rec.ResponseWriter.WriteHeader(status)
}

func (rec *statusCodeRecorder) Write(b []byte) (int, error) {
	if rec.Status == 0 {
		rec.Status = http.StatusOK
	}

	return rec.ResponseWriter.Write(b)
}
