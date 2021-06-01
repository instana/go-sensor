package instana

import (
	"net/http"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// HttpEntrySpan is a wrapper for an ot.Span. It can be used to init span for incoming HTTP requests.
// HttpEntrySpan has methods to collect meta information from the request and response objects.
type HttpEntrySpan struct {
	span                   ot.Span
	collectableHTTPHeaders []string
	secrets                Matcher
	collectedHeaders       map[string]string
}

// NewHttpEntrySpan returns instance of *HttpEntrySpan with tags created from the *http.Request and pathTemplate.
func NewHttpEntrySpan(req *http.Request, sensor *Sensor, pathTemplate string) *HttpEntrySpan {
	s := &HttpEntrySpan{
		collectedHeaders: map[string]string{},
	}

	s.initHttpSpan(req, sensor, pathTemplate)
	s.collectRequestHeadersAndParams(req)
	return s
}

// CollectPanicInformation collects data from the recovered value (function recover()) in case of the panic.
func (ht *HttpEntrySpan) CollectPanicInformation(err interface{}) {
	if e, ok := err.(error); ok {
		ht.span.SetTag("http.error", e.Error())
		ht.span.LogFields(otlog.Error(e))
	} else {
		ht.span.SetTag("http.error", err)
		ht.span.LogFields(otlog.Object("error", err))
	}

	ht.span.SetTag(string(ext.HTTPStatusCode), http.StatusInternalServerError)
}

// CollectResponseHeaders collects response headers.
func (ht *HttpEntrySpan) CollectResponseHeaders(w http.ResponseWriter) {
	for _, h := range ht.collectableHTTPHeaders {
		if v := w.Header().Get(h); v != "" {
			ht.collectedHeaders[h] = v
		}
	}
}

// CollectResponseStatus sets the status. It stores error status text in case of error.
func (ht *HttpEntrySpan) CollectResponseStatus(status int) {
	if status > 0 {
		if status >= http.StatusInternalServerError {
			statusText := http.StatusText(status)

			ht.span.SetTag("http.error", statusText)
			ht.span.LogFields(otlog.Object("error", statusText))
		}

		ht.span.SetTag("http.status", status)
	}
}

// Inject tracing context into the ResponseWriter.
func (ht *HttpEntrySpan) Inject(w http.ResponseWriter) error {
	return ht.span.Tracer().Inject(ht.span.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(w.Header()))
}

// RequestWithContext returns *http.Request with context that contains the span.
func (ht *HttpEntrySpan) RequestWithContext(req *http.Request) *http.Request {
	originalCtx := req.Context()
	ctxWithSpan := ContextWithSpan(originalCtx, ht.span)

	return req.WithContext(ctxWithSpan)
}

// Finish writes collected headers and calls Finish() on the wrapped span. Must be the last call made to any span instance.
func (ht *HttpEntrySpan) Finish() {
	ht.writeHeaders()
	ht.span.Finish()
}

func (ht *HttpEntrySpan) initHttpSpan(req *http.Request, sensor *Sensor, pathTemplate string) {
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

	ht.span = tracer.StartSpan("g.http", opts...)
}

func (ht *HttpEntrySpan) collectRequestHeadersAndParams(req *http.Request) {
	params := collectHTTPParams(req, ht.secrets)
	if len(params) > 0 {
		ht.span.SetTag("http.params", params.Encode())
	}

	for _, h := range ht.collectableHTTPHeaders {
		if v := req.Header.Get(h); v != "" {
			ht.collectedHeaders[h] = v
		}
	}
}

func (ht *HttpEntrySpan) writeHeaders() {
	if len(ht.collectedHeaders) > 0 {
		ht.span.SetTag("http.header", ht.collectedHeaders)
	}
}
