package instana

import (
	"net/http"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

type StatusReader interface {
	Status() int
}

type HttpSpan struct {
	span                   ot.Span
	tracer                 ot.Tracer
	collectableHTTPHeaders []string
	secrets                Matcher
	collectedHeaders       map[string]string
}

func NewHttpSpan(req *http.Request, sensor *Sensor, pathTemplate string) *HttpSpan {
	s := &HttpSpan{
		collectedHeaders: map[string]string{},
	}

	s.initHttpSpan(req, sensor, pathTemplate)
	s.collectRequestHeadersAndParams(req)
	return s
}

func (ht *HttpSpan) CollectPanicInformation(err interface{}) {
	if e, ok := err.(error); ok {
		ht.span.SetTag("http.error", e.Error())
		ht.span.LogFields(otlog.Error(e))
	} else {
		ht.span.SetTag("http.error", err)
		ht.span.LogFields(otlog.Object("error", err))
	}

	ht.span.SetTag(string(ext.HTTPStatusCode), http.StatusInternalServerError)
}

func (ht *HttpSpan) CollectResponseHeaders(w http.ResponseWriter) {
	// collect response headers
	for _, h := range ht.collectableHTTPHeaders {
		if v := w.Header().Get(h); v != "" {
			ht.collectedHeaders[h] = v
		}
	}
}

func (ht *HttpSpan) CollectResponseStatus(r StatusReader) {
	if r.Status() > 0 {
		if r.Status() >= http.StatusInternalServerError {
			statusText := http.StatusText(r.Status())

			ht.span.SetTag("http.error", statusText)
			ht.span.LogFields(otlog.Object("error", statusText))
		}

		ht.span.SetTag("http.status", r.Status())
	}
}

func (ht *HttpSpan) collectRequestHeadersAndParams(req *http.Request) {
	params := collectHTTPParams(req, ht.secrets)
	if len(params) > 0 {
		ht.span.SetTag("http.params", params.Encode())
	}

	// collect request headers
	for _, h := range ht.collectableHTTPHeaders {
		if v := req.Header.Get(h); v != "" {
			ht.collectedHeaders[h] = v
		}
	}
}

func (ht *HttpSpan) Finish() {
	ht.writeHeaders()
	ht.span.Finish()
}

func (ht *HttpSpan) Inject(w http.ResponseWriter) error {
	return ht.tracer.Inject(ht.span.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(w.Header()))
}

func (ht *HttpSpan) initHttpSpan(req *http.Request, sensor *Sensor, pathTemplate string) {
	opts := []ot.StartSpanOption{
		ext.SpanKindRPCServer,
		ot.Tags{
			"http.host":     req.Host,
			"http.method":   req.Method,
			"http.protocol": req.URL.Scheme,
			"http.path":     req.URL.Path,
		},
	}

	tracer := sensor.Tracer()
	if ps, ok := SpanFromContext(req.Context()); ok {
		tracer = ps.Tracer()
		opts = append(opts, ot.ChildOf(ps.Context()))
	}

	if t, ok := tracer.(Tracer); ok {
		opts := t.Options()
		ht.collectableHTTPHeaders = opts.CollectableHTTPHeaders
		ht.secrets = opts.Secrets
	}

	wireContext, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))
	switch err {
	case nil:
		opts = append(opts, ext.RPCServerOption(wireContext))
	case ot.ErrSpanContextNotFound:
		sensor.Logger().Debug("no span context provided with ", req.Method, " ", req.URL.Path)
	case ot.ErrUnsupportedFormat:
		sensor.Logger().Info("unsupported span context format provided with ", req.Method, " ", req.URL.Path)
	default:
		sensor.Logger().Warn("failed to extract span context from the request:", err)
	}

	if req.Header.Get(FieldSynthetic) == "1" {
		opts = append(opts, syntheticCall())
	}

	if pathTemplate != "" && req.URL.Path != pathTemplate {
		opts = append(opts, ot.Tag{Key: "http.path_tpl", Value: pathTemplate})
	}

	span := tracer.StartSpan("g.http", opts...)

	ht.tracer = tracer
	ht.span = span
}

func (ht *HttpSpan) writeHeaders() {
	if len(ht.collectedHeaders) > 0 {
		ht.span.SetTag("http.header", ht.collectedHeaders)
	}
}

func (ht *HttpSpan) RequestWithContext(req *http.Request) *http.Request {
	originalCtx := req.Context()
	ctxWithSpan := ContextWithSpan(originalCtx, ht.span)

	return req.WithContext(ctxWithSpan)
}
