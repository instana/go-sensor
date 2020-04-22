package instana_test

import (
	"net/http"
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracer_Inject_Extract_HTTPHeaders(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test-span")
	sp.SetBaggageItem("Foo", "bar")

	headers := http.Header{}

	require.NoError(t, tracer.Inject(sp.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers)))
	sp.Finish()

	sc, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers))
	require.NoError(t, err)

	assert.Equal(t, sp.Context(), sc)
}

func TestTracer_Inject_HTTPHeaders(t *testing.T) {
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

func TestTracer_Extract_HTTPHeaders(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test-span")
	sp.SetBaggageItem("foo", "bar")

	headers := http.Header{
		"Authorization":   {"Basic 123"},
		"x-instana-t":     {"1314"},
		"X-INSTANA-S":     {"2435"},
		"X-Instana-L":     {"1"},
		"X-Instana-B-foo": {"bar"},
	}

	sc, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers))
	require.NoError(t, err)

	assert.Equal(t, instana.SpanContext{
		TraceID: 0x1314,
		SpanID:  0x2435,
		Baggage: map[string]string{
			"foo": "bar",
		},
	}, sc)
}

func TestTracer_Extract_HTTPHeaders_NoContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test-span")
	sp.SetBaggageItem("foo", "bar")

	headers := http.Header{
		"Authorization": {"Basic 123"},
	}

	_, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers))
	assert.Equal(t, ot.ErrSpanContextNotFound, err)
}

func TestTracer_Extract_HTTPHeaders_CorruptedContext(t *testing.T) {
	examples := map[string]http.Header{
		"missing trace id": {
			"X-INSTANA-S": {"1314"},
			"X-Instana-L": {"1"},
		},
		"malformed trace id": {
			"x-instana-t": {"wrong"},
			"X-INSTANA-S": {"1314"},
			"X-Instana-L": {"1"},
		},
		"missing span id": {
			"x-instana-t": {"1314"},
			"X-Instana-L": {"1"},
		},
		"malformed span id": {
			"x-instana-t": {"1314"},
			"X-INSTANA-S": {"wrong"},
			"X-Instana-L": {"1"},
		},
	}

	for name, headers := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

			sp := tracer.StartSpan("test-span")
			sp.SetBaggageItem("foo", "bar")

			_, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers))
			assert.Equal(t, ot.ErrSpanContextCorrupted, err)
		})
	}
}

func TestTracer_Inject_Extract_TextMap(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test-span")
	sp.SetBaggageItem("foo", "bar")

	carrier := make(map[string]string)

	require.NoError(t, tracer.Inject(sp.Context(), ot.TextMap, ot.TextMapCarrier(carrier)))
	sp.Finish()

	sc, err := tracer.Extract(ot.TextMap, ot.TextMapCarrier(carrier))
	require.NoError(t, err)

	assert.Equal(t, sp.Context(), sc)
}

func TestTracer_Inject_TextMap_AddValues(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test-span")
	sp.SetBaggageItem("foo", "bar")

	carrier := map[string]string{
		"key1": "value1",
	}

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
}

func TestTracer_Inject_TextMap_UpdateValues(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test-span")
	sp.SetBaggageItem("foo", "bar")

	carrier := map[string]string{
		"key1":            "value1",
		"x-instana-t":     "1314",
		"X-INSTANA-S":     "1314",
		"X-Instana-L":     "1",
		"X-INSTANA-b-foo": "hello",
	}

	require.NoError(t, tracer.Inject(sp.Context(), ot.TextMap, ot.TextMapCarrier(carrier)))
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	require.Equal(t, map[string]string{
		"x-instana-t":     instana.FormatID(span.TraceID),
		"X-INSTANA-S":     instana.FormatID(span.SpanID),
		"X-Instana-L":     "1",
		"X-INSTANA-b-foo": "bar",
		"key1":            "value1",
	}, carrier)
}

func TestTracer_Extract_TextMap(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test-span")
	sp.SetBaggageItem("foo", "bar")

	carrier := map[string]string{
		"key1":            "value1",
		"x-instana-t":     "1314",
		"x-instana-s":     "2435",
		"x-instana-l":     "1",
		"x-instana-b-foo": "bar",
	}

	sc, err := tracer.Extract(ot.TextMap, ot.TextMapCarrier(carrier))
	require.NoError(t, err)

	assert.Equal(t, instana.SpanContext{
		TraceID: 0x1314,
		SpanID:  0x2435,
		Baggage: map[string]string{
			"foo": "bar",
		},
	}, sc)
}

func TestTracer_Extract_TextMap_NoContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test-span")
	sp.SetBaggageItem("foo", "bar")

	carrier := map[string]string{
		"key1": "value1",
	}

	_, err := tracer.Extract(ot.TextMap, ot.TextMapCarrier(carrier))
	assert.Equal(t, ot.ErrSpanContextNotFound, err)
}

func TestTracer_Extract_TextMap_CorruptedContext(t *testing.T) {
	examples := map[string]map[string]string{
		"missing trace id": {
			"x-instana-s": "1314",
			"x-instana-l": "1",
		},
		"malformed trace id": {
			"x-instana-t": "wrong",
			"x-instana-s": "1314",
			"x-instana-l": "1",
		},
		"missing span id": {
			"x-instana-t": "1314",
			"x-instana-l": "1",
		},
		"malformed span id": {
			"x-instana-t": "1314",
			"x-instana-s": "wrong",
			"x-instana-l": "1",
		},
	}

	for name, carrier := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

			sp := tracer.StartSpan("test-span")
			sp.SetBaggageItem("foo", "bar")

			_, err := tracer.Extract(ot.TextMap, ot.TextMapCarrier(carrier))
			assert.Equal(t, ot.ErrSpanContextCorrupted, err)
		})
	}
}
