package instana_test

import (
	"net/http"
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropagation_Inject_HTTPHeadersCarrier(t *testing.T) {
	examples := map[string]http.Header{
		"add headers": {
			"Authorization": {"Basic 123"},
		},
		"update headers": {
			"Authorization":   {"Basic 123"},
			"x-instana-t":     {"1314"},
			"X-INSTANA-S":     {"1314"},
			"X-Instana-L":     {"1"},
			"X-Instana-B-foo": {"hello"},
		},
	}

	for name, headers := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

			sp := tracer.StartSpan("test-span")
			sp.SetBaggageItem("foo", "bar")

			require.NoError(t, tracer.Inject(sp.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers)))
			sp.Finish()

			require.Len(t, headers, 5)

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			span := spans[0]
			assert.Equal(t, instana.FormatID(span.TraceID), headers.Get("X-Instana-T"))
			assert.Equal(t, instana.FormatID(span.SpanID), headers.Get("X-Instana-S"))
			assert.Equal(t, "1", headers.Get("X-Instana-L"))
			assert.Equal(t, "bar", headers.Get("X-Instana-B-foo"))
			assert.Equal(t, "Basic 123", headers.Get("Authorization"))
		})
	}
}

func TestPropagation_TextMap(t *testing.T) {
	examples := map[string]map[string]string{
		"add values": {
			"key1": "value1",
		},
		"update values": {
			"key1":            "value1",
			"x-instana-t":     "1314",
			"x-instana-s":     "1314",
			"x-instana-l":     "1",
			"x-instana-b-foo": "hello",
		},
	}

	for name, carrier := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

			sp := tracer.StartSpan("test-span")
			sp.SetBaggageItem("foo", "bar")

			require.NoError(t, tracer.Inject(sp.Context(), ot.TextMap, ot.TextMapCarrier(carrier)))
			sp.Finish()

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			span := spans[0]
			require.Equal(t, map[string]string{
				"x-instana-t":     instana.FormatID(span.TraceID),
				"x-instana-s":     instana.FormatID(span.SpanID),
				"x-instana-l":     "1",
				"x-instana-b-foo": "bar",
				"key1":            "value1",
			}, carrier)
		})
	}
}
