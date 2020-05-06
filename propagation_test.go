package instana_test

import (
	"net/http"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/w3ctrace"
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

			sc := instana.SpanContext{
				TraceID: 0x2435,
				SpanID:  0x3546,
				ForeignParent: w3ctrace.Context{
					RawParent: "w3cparent",
					RawState:  "w3cstate",
				},
				Baggage: map[string]string{
					"foo": "bar",
				},
			}

			require.NoError(t, tracer.Inject(sc, ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers)))

			// Instana trace context
			assert.Equal(t, "2435", headers.Get("X-Instana-T"))
			assert.Equal(t, "3546", headers.Get("X-Instana-S"))
			assert.Equal(t, "1", headers.Get("X-Instana-L"))
			assert.Equal(t, "bar", headers.Get("X-Instana-B-foo"))
			// W3C trace context
			assert.Equal(t, "w3cparent", headers.Get(w3ctrace.TraceParentHeader))
			assert.Equal(t, "w3cstate", headers.Get(w3ctrace.TraceStateHeader))
			// Original headers
			assert.Equal(t, "Basic 123", headers.Get("Authorization"))

			assert.Len(t, headers, 7)
		})
	}
}

func TestTracer_Inject_HTTPHeaders_SuppressedTracing(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	headers := http.Header{
		"Authorization": {"Basic 123"},
		"x-instana-t":   {"1314"},
		"X-INSTANA-S":   {"1314"},
		"X-Instana-L":   {"1"},
	}

	sc := instana.SpanContext{
		TraceID:    0x2435,
		SpanID:     0x3546,
		Suppressed: true,
	}

	require.NoError(t, tracer.Inject(sc, ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers)))

	assert.Equal(t, "2435", headers.Get("X-Instana-T"))
	assert.Equal(t, "3546", headers.Get("X-Instana-S"))
	assert.Equal(t, "0", headers.Get("X-Instana-L"))
	assert.Equal(t, "Basic 123", headers.Get("Authorization"))
}

func TestTracer_Extract_HTTPHeaders(t *testing.T) {
	examples := map[string]struct {
		Headers  map[string]string
		Expected instana.SpanContext
	}{
		"tracing enabled": {
			Headers: map[string]string{
				"Authorization":   "Basic 123",
				"x-instana-t":     "1314",
				"X-INSTANA-S":     "2435",
				"X-Instana-L":     "1",
				"X-Instana-B-Foo": "bar",
			},
			Expected: instana.SpanContext{
				TraceID: 0x1314,
				SpanID:  0x2435,
				Baggage: map[string]string{
					"Foo": "bar",
				},
			},
		},
		"tracing disabled": {
			Headers: map[string]string{
				"Authorization": "Basic 123",
				"x-instana-t":   "1314",
				"X-INSTANA-S":   "2435",
				"X-Instana-L":   "0",
			},
			Expected: instana.SpanContext{
				TraceID:    0x1314,
				SpanID:     0x2435,
				Suppressed: true,
				Baggage:    map[string]string{},
			},
		},
		"w3c trace context": {
			Headers: map[string]string{
				"x-instana-t": "1314",
				"X-INSTANA-S": "2435",
				"X-Instana-L": "1",
				"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				"tracestate":  "rojo=00f067aa0ba902b7",
			},
			Expected: instana.SpanContext{
				TraceID: 0x1314,
				SpanID:  0x2435,
				Baggage: map[string]string{},
				ForeignParent: w3ctrace.Context{
					RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					RawState:  "rojo=00f067aa0ba902b7",
				},
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

			headers := http.Header{}
			for k, v := range example.Headers {
				headers.Set(k, v)
			}

			sc, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers))
			require.NoError(t, err)

			assert.Equal(t, example.Expected, sc)
		})
	}
}

func TestTracer_Extract_HTTPHeaders_NoContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

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

	sc := instana.SpanContext{
		TraceID: 0x2435,
		SpanID:  0x3546,
		Baggage: map[string]string{
			"foo": "bar",
		},
	}

	carrier := map[string]string{
		"key1": "value1",
	}

	require.NoError(t, tracer.Inject(sc, ot.TextMap, ot.TextMapCarrier(carrier)))

	assert.Equal(t, map[string]string{
		"x-instana-t":     "2435",
		"x-instana-s":     "3546",
		"x-instana-l":     "1",
		"x-instana-b-foo": "bar",
		"key1":            "value1",
	}, carrier)
}

func TestTracer_Inject_TextMap_UpdateValues(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sc := instana.SpanContext{
		TraceID: 0x2435,
		SpanID:  0x3546,
		Baggage: map[string]string{
			"foo": "bar",
		},
	}

	carrier := map[string]string{
		"key1":            "value1",
		"x-instana-t":     "1314",
		"X-INSTANA-S":     "1314",
		"X-Instana-L":     "1",
		"X-INSTANA-b-foo": "hello",
	}

	require.NoError(t, tracer.Inject(sc, ot.TextMap, ot.TextMapCarrier(carrier)))

	assert.Equal(t, map[string]string{
		"x-instana-t":     "2435",
		"X-INSTANA-S":     "3546",
		"X-Instana-L":     "1",
		"X-INSTANA-b-foo": "bar",
		"key1":            "value1",
	}, carrier)
}

func TestTracer_Inject_TextMap_SuppressedTracing(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sc := instana.SpanContext{
		TraceID:    0x2435,
		SpanID:     0x3546,
		Suppressed: true,
	}

	carrier := map[string]string{
		"key1":        "value1",
		"x-instana-t": "1314",
		"X-INSTANA-S": "1314",
		"X-Instana-L": "1",
	}

	require.NoError(t, tracer.Inject(sc, ot.TextMap, ot.TextMapCarrier(carrier)))

	assert.Equal(t, map[string]string{
		"x-instana-t": "2435",
		"X-INSTANA-S": "3546",
		"X-Instana-L": "0",
		"key1":        "value1",
	}, carrier)
}

func TestTracer_Extract_TextMap(t *testing.T) {
	examples := map[string]struct {
		Carrier  map[string]string
		Expected instana.SpanContext
	}{
		"tracing enabled": {
			Carrier: map[string]string{
				"Authorization":   "Basic 123",
				"x-instana-t":     "1314",
				"X-INSTANA-S":     "2435",
				"X-Instana-L":     "1",
				"X-Instana-B-Foo": "bar",
			},
			Expected: instana.SpanContext{
				TraceID: 0x1314,
				SpanID:  0x2435,
				Baggage: map[string]string{
					"Foo": "bar",
				},
			},
		},
		"tracing disabled": {
			Carrier: map[string]string{
				"Authorization": "Basic 123",
				"x-instana-t":   "1314",
				"X-INSTANA-S":   "2435",
				"X-Instana-L":   "0",
			},
			Expected: instana.SpanContext{
				TraceID:    0x1314,
				SpanID:     0x2435,
				Suppressed: true,
				Baggage:    map[string]string{},
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

			sc, err := tracer.Extract(ot.TextMap, ot.TextMapCarrier(example.Carrier))
			require.NoError(t, err)

			assert.Equal(t, example.Expected, sc)
		})
	}
}

func TestTracer_Extract_TextMap_NoContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

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

			_, err := tracer.Extract(ot.TextMap, ot.TextMapCarrier(carrier))
			assert.Equal(t, ot.ErrSpanContextCorrupted, err)
		})
	}
}
