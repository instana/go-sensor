// (c) Copyright IBM Corp. 2022

package instana

import (
	ot "github.com/opentracing/opentracing-go"
	"net/http"
	"net/url"
	"sync"
)

const maxDelayedSpans = 500

var delayed = delayedSpans{
	mu:    sync.Mutex{},
	spans: make(map[int64]delayedSpan),
}

type delayedSpan struct {
	span            ot.Span
	requestHeaders  http.Header
	responseHeaders http.Header
	httpParams      url.Values
}

type delayedSpans struct {
	mu    sync.Mutex
	spans map[int64]delayedSpan
}

func (ds *delayedSpans) append(dSpan delayedSpan) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if len(ds.spans) <= maxDelayedSpans {
		if s, ok := dSpan.span.(*spanS); ok {

			if dSpan.span == nil {
				return
			}

			spanID := s.context.SpanID
			ds.spans[spanID] = dSpan
		}
	}
}

func (ds *delayedSpans) flush() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if sensor.Agent().Ready() && len(ds.spans) > 0 {
		for sid, s := range ds.spans {
			if t, ok := s.span.Tracer().(Tracer); ok {
				opts := t.Options()

				setHeaders(s.requestHeaders, s.responseHeaders, opts.CollectableHTTPHeaders, s.span)

				params := collectHTTPParams(s.httpParams, opts.Secrets)
				if len(params) > 0 {
					s.span.SetTag("http.params", params.Encode())
				}

				if sp, ok := s.span.(*spanS); ok {
					sp.tracer.recorder.RecordSpan(sp)
					sp.sendOpenTracingLogRecords()
				}
			}

			delete(ds.spans, sid)
		}

	}
}

func (ds *delayedSpans) setSpan(s *spanS) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	spanID := s.context.SpanID

	if _, ok := ds.spans[spanID]; ok {
		return
	}

	if len(ds.spans) <= maxDelayedSpans {
		ds.spans[spanID] = delayedSpan{
			span: s,
		}
	}
}
