// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

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

func TestTracer_Inject_HTTPHeaders(t *testing.T) {
	examples := map[string]struct {
		SpanContext instana.SpanContext
		Headers     http.Header
		Expected    http.Header
	}{
		"no trace context": {
			SpanContext: instana.SpanContext{
				TraceIDHi: 0x1,
				TraceID:   0x2435,
				SpanID:    0x3546,
				Baggage: map[string]string{
					"foo": "bar",
				},
			},
			Headers: http.Header{
				"Authorization": {"Basic 123"},
			},
			Expected: http.Header{
				"Authorization":   {"Basic 123"},
				"X-Instana-T":     {"0000000000002435"},
				"X-Instana-S":     {"0000000000003546"},
				"X-Instana-L":     {"1"},
				"X-Instana-B-Foo": {"bar"},
				"Traceparent":     {"00-00000000000000010000000000002435-0000000000003546-01"},
				"Tracestate":      {"in=0000000000002435;0000000000003546"},
				"Server-Timing":   {"intid;desc=0000000000002435"},
			},
		},
		"with instana trace": {
			SpanContext: instana.SpanContext{
				TraceIDHi: 0x1,
				TraceID:   0x2435,
				SpanID:    0x3546,
				Baggage: map[string]string{
					"foo": "bar",
				},
			},
			Headers: http.Header{
				"Authorization":   {"Basic 123"},
				"x-instana-t":     {"0000000000001314"},
				"X-INSTANA-S":     {"0000000000001314"},
				"X-Instana-L":     {"1"},
				"X-Instana-B-foo": {"hello"},
			},
			Expected: http.Header{
				"Authorization":   {"Basic 123"},
				"X-Instana-T":     {"0000000000002435"},
				"X-Instana-S":     {"0000000000003546"},
				"X-Instana-L":     {"1"},
				"X-Instana-B-Foo": {"bar"},
				"Traceparent":     {"00-00000000000000010000000000002435-0000000000003546-01"},
				"Tracestate":      {"in=0000000000002435;0000000000003546"},
				"Server-Timing":   {"intid;desc=0000000000002435"},
			},
		},
		"with instana trace suppressed": {
			SpanContext: instana.SpanContext{
				TraceIDHi:  0x1,
				TraceID:    0x2435,
				SpanID:     0x3546,
				Suppressed: true,
			},
			Headers: http.Header{
				"Authorization": {"Basic 123"},
			},
			Expected: http.Header{
				"Authorization": {"Basic 123"},
				"X-Instana-L":   {"0"},
				"Traceparent":   {"00-00000000000000010000000000002435-0000000000003546-00"},
				"Server-Timing": {"intid;desc=0000000000002435"},
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
			defer instana.ShutdownSensor()

			require.NoError(t, tracer.Inject(example.SpanContext, ot.HTTPHeaders, ot.HTTPHeadersCarrier(example.Headers)))
			assert.Equal(t, example.Expected, example.Headers)
		})
	}
}

func TestTracer_Inject_HTTPHeaders_W3CTraceContext(t *testing.T) {
	examples := map[string]struct {
		SpanContext instana.SpanContext
		Expected    http.Header
	}{
		"instana trace suppressed, no w3c trace": {
			SpanContext: instana.SpanContext{
				TraceIDHi:  0x01,
				TraceID:    0x2435,
				SpanID:     0x3546,
				Suppressed: true,
			},
			Expected: http.Header{
				"X-Instana-L":   {"0"},
				"Traceparent":   {"00-00000000000000010000000000002435-0000000000003546-00"},
				"Server-Timing": {"intid;desc=0000000000002435"},
			},
		},
		"instana trace suppressed, w3c trace not sampled": {
			SpanContext: instana.SpanContext{
				TraceIDHi: 0x01,
				TraceID:   0x2435,
				SpanID:    0x3546,
				W3CContext: w3ctrace.Context{
					RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00",
					RawState:  "rojo=00f067aa0ba902b7",
				},
				Suppressed: true,
			},
			Expected: http.Header{
				"X-Instana-L":   {"0"},
				"Traceparent":   {"00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000003546-00"},
				"Tracestate":    {"rojo=00f067aa0ba902b7"},
				"Server-Timing": {"intid;desc=0000000000002435"},
			},
		},
		"instana trace suppressed, w3c trace sampled": {
			SpanContext: instana.SpanContext{
				TraceIDHi: 0x01,
				TraceID:   0x2435,
				SpanID:    0x3546,
				W3CContext: w3ctrace.Context{
					RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					RawState:  "rojo=00f067aa0ba902b7",
				},
				Suppressed: true,
			},
			Expected: http.Header{
				"X-Instana-L":   {"0"},
				"Traceparent":   {"00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000003546-00"},
				"Tracestate":    {"rojo=00f067aa0ba902b7"},
				"Server-Timing": {"intid;desc=0000000000002435"},
			},
		},
		"instana trace, no w3c trace": {
			SpanContext: instana.SpanContext{
				TraceIDHi: 0x01,
				TraceID:   0x2435,
				SpanID:    0x3546,
			},
			Expected: http.Header{
				"X-Instana-T":   {"0000000000002435"},
				"X-Instana-S":   {"0000000000003546"},
				"X-Instana-L":   {"1"},
				"Traceparent":   {"00-00000000000000010000000000002435-0000000000003546-01"},
				"Tracestate":    {"in=0000000000002435;0000000000003546"},
				"Server-Timing": {"intid;desc=0000000000002435"},
			},
		},
		"instana trace, w3c trace not sampled": {
			SpanContext: instana.SpanContext{
				TraceIDHi: 0x01,
				TraceID:   0x2435,
				SpanID:    0x3546,
				W3CContext: w3ctrace.Context{
					RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00",
					RawState:  "rojo=00f067aa0ba902b7",
				},
			},
			Expected: http.Header{
				"X-Instana-T":   {"0000000000002435"},
				"X-Instana-S":   {"0000000000003546"},
				"X-Instana-L":   {"1"},
				"Traceparent":   {"00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000003546-01"},
				"Tracestate":    {"in=0000000000002435;0000000000003546,rojo=00f067aa0ba902b7"},
				"Server-Timing": {"intid;desc=0000000000002435"},
			},
		},
		"instana trace, w3c trace": {
			SpanContext: instana.SpanContext{
				TraceIDHi: 0x01,
				TraceID:   0x2435,
				SpanID:    0x3546,
				W3CContext: w3ctrace.Context{
					RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					RawState:  "rojo=00f067aa0ba902b7",
				},
			},
			Expected: http.Header{
				"X-Instana-T":   {"0000000000002435"},
				"X-Instana-S":   {"0000000000003546"},
				"X-Instana-L":   {"1"},
				"Traceparent":   {"00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000003546-01"},
				"Tracestate":    {"in=0000000000002435;0000000000003546,rojo=00f067aa0ba902b7"},
				"Server-Timing": {"intid;desc=0000000000002435"},
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
			defer instana.ShutdownSensor()
			headers := http.Header{}

			require.NoError(t, tracer.Inject(example.SpanContext, ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers)))
			assert.Equal(t, example.Expected, headers)
		})
	}
}

func TestTracer_Inject_HTTPHeaders_SuppressedTracing(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	headers := http.Header{
		"Authorization": {"Basic 123"},
		"x-instana-t":   {"0000000000001314"},
		"X-INSTANA-S":   {"0000000000001314"},
		"X-Instana-L":   {"1"},
	}

	sc := instana.SpanContext{
		TraceIDHi:  0x1,
		TraceID:    0x2435,
		SpanID:     0x3546,
		Suppressed: true,
	}

	require.NoError(t, tracer.Inject(sc, ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers)))

	assert.Empty(t, headers.Get("X-Instana-T"))
	assert.Empty(t, headers.Get("X-Instana-S"))
	assert.Equal(t, "0", headers.Get("X-Instana-L"))
	assert.Equal(t, "Basic 123", headers.Get("Authorization"))
	assert.Equal(t, "intid;desc=0000000000002435", headers.Get("Server-Timing"))
}

func TestTracer_Inject_HTTPHeaders_WithExistingServerTiming(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	headers := http.Header{
		"x-instana-t":   {"0000000000001314"},
		"X-INSTANA-S":   {"0000000000001314"},
		"X-Instana-L":   {"1"},
		"Server-Timing": {"db;dur=53, app;dur=47.2", `cache;desc="Cache Read";dur=23.2`},
	}

	sc := instana.SpanContext{
		TraceID:    0x2435,
		SpanID:     0x3546,
		Suppressed: true,
	}

	require.NoError(t, tracer.Inject(sc, ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers)))
	assert.Equal(t, `db;dur=53, app;dur=47.2, cache;desc="Cache Read";dur=23.2, intid;desc=0000000000002435`, headers.Get("Server-Timing"))
}

func TestTracer_Extract_HTTPHeaders(t *testing.T) {
	examples := map[string]struct {
		Headers  map[string]string
		Expected instana.SpanContext
	}{
		"tracing enabled": {
			Headers: map[string]string{
				"Authorization":   "Basic 123",
				"x-instana-t":     "0000000000000000000000010000000000001314",
				"X-INSTANA-S":     "0000000000002435",
				"X-Instana-L":     "1",
				"X-Instana-B-Foo": "bar",
			},
			Expected: instana.SpanContext{
				TraceIDHi: 0x1,
				TraceID:   0x1314,
				SpanID:    0x2435,
				Baggage: map[string]string{
					"Foo": "bar",
				},
			},
		},
		"tracing disabled": {
			Headers: map[string]string{
				"Authorization": "Basic 123",
				"x-instana-t":   "0000000000000000000000010000000000001314",
				"X-INSTANA-S":   "0000000000002435",
				"X-Instana-L":   "0",
			},
			Expected: instana.SpanContext{
				TraceIDHi:  0x1,
				TraceID:    0x1314,
				SpanID:     0x2435,
				Suppressed: true,
				Baggage:    map[string]string{},
			},
		},
		"tracing disabled, with correlation data": {
			Headers: map[string]string{
				"Authorization": "Basic 123",
				"x-instana-t":   "10000000000001314",
				"X-INSTANA-S":   "2435",
				"X-Instana-L":   "0,correlationType=web;correlationId=1234",
			},
			Expected: instana.SpanContext{
				TraceIDHi:  0x1,
				TraceID:    0x1314,
				SpanID:     0x2435,
				Suppressed: true,
				Baggage:    map[string]string{},
			},
		},
		"tracing disabled, no trace context": {
			Headers: map[string]string{
				"Authorization": "Basic 123",
				"X-Instana-L":   "0",
			},
			Expected: instana.SpanContext{
				Suppressed: true,
				Baggage:    map[string]string{},
			},
		},
		"w3c trace context, with instana headers": {
			Headers: map[string]string{
				"x-instana-t": "10000000000001314",
				"X-INSTANA-S": "2435",
				"X-Instana-L": "1",
				"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				"tracestate":  "rojo=00f067aa0ba902b7",
			},
			Expected: instana.SpanContext{
				TraceIDHi: 0x1,
				TraceID:   0x1314,
				SpanID:    0x2435,
				Baggage:   map[string]string{},
				W3CContext: w3ctrace.Context{
					RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					RawState:  "rojo=00f067aa0ba902b7",
				},
			},
		},
		"w3c trace context, no instana headers": {
			Headers: map[string]string{
				"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				"tracestate":  "in=10000000000001314;2435,rojo=00f067aa0ba902b7",
			},
			Expected: instana.SpanContext{
				Baggage: map[string]string{},
				W3CContext: w3ctrace.Context{
					RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					RawState:  "in=10000000000001314;2435,rojo=00f067aa0ba902b7",
				},
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
			defer instana.ShutdownSensor()

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

func TestTracer_Extract_HTTPHeaders_WithEUMCorrelation(t *testing.T) {
	examples := map[string]struct {
		Headers  map[string]string
		Expected instana.SpanContext
	}{
		"tracing enabled, no instana headers": {
			Headers: map[string]string{
				"X-Instana-L": "1,correlationType=web;correlationId=1234",
			},
		},
		"tracing enabled, with instana headers": {
			Headers: map[string]string{
				"X-Instana-T": "0000000000002435",
				"X-Instana-S": "0000000000003546",
				"X-Instana-L": "1,correlationType=web;correlationId=1234",
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
			defer instana.ShutdownSensor()

			headers := http.Header{}
			for k, v := range example.Headers {
				headers.Set(k, v)
			}

			sc, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers))
			require.NoError(t, err)

			spanContext := sc.(instana.SpanContext)

			assert.EqualValues(t, 0, spanContext.TraceID)
			assert.EqualValues(t, 0, spanContext.SpanID)
			assert.Empty(t, spanContext.ParentID)
			assert.Equal(t, instana.EUMCorrelationData{ID: "1234", Type: "web"}, spanContext.Correlation)
		})
	}
}

func TestTracer_Extract_HTTPHeaders_NoContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

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
		"missing span id": {
			"x-instana-t": {"1314"},
			"X-Instana-L": {"1"},
		},
		"malformed trace id": {
			"x-instana-t": {"wrong"},
			"X-INSTANA-S": {"1314"},
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
			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
			defer instana.ShutdownSensor()

			_, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers))
			assert.Equal(t, ot.ErrSpanContextCorrupted, err)
		})
	}
}

func TestTracer_Inject_TextMap_AddValues(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	sc := instana.SpanContext{
		TraceIDHi: 0x1,
		TraceID:   0x2435,
		SpanID:    0x3546,
		Baggage: map[string]string{
			"foo": "bar",
		},
	}

	carrier := map[string]string{
		"key1": "value1",
	}

	require.NoError(t, tracer.Inject(sc, ot.TextMap, ot.TextMapCarrier(carrier)))

	assert.Equal(t, map[string]string{
		"x-instana-t":     "0000000000002435",
		"x-instana-s":     "0000000000003546",
		"x-instana-l":     "1",
		"x-instana-b-foo": "bar",
		"key1":            "value1",
	}, carrier)
}

func TestTracer_Inject_TextMap_UpdateValues(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	sc := instana.SpanContext{
		TraceIDHi: 0x1,
		TraceID:   0x2435,
		SpanID:    0x3546,
		Baggage: map[string]string{
			"foo": "bar",
		},
	}

	carrier := map[string]string{
		"key1":            "value1",
		"x-instana-t":     "0000000000001314",
		"X-INSTANA-S":     "0000000000001314",
		"X-Instana-L":     "1",
		"X-INSTANA-b-foo": "hello",
	}

	require.NoError(t, tracer.Inject(sc, ot.TextMap, ot.TextMapCarrier(carrier)))

	assert.Equal(t, map[string]string{
		"x-instana-t":     "0000000000002435",
		"X-INSTANA-S":     "0000000000003546",
		"X-Instana-L":     "1",
		"X-INSTANA-b-foo": "bar",
		"key1":            "value1",
	}, carrier)
}

func TestTracer_Inject_TextMap_SuppressedTracing(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	sc := instana.SpanContext{
		TraceIDHi:  0x1,
		TraceID:    0x2435,
		SpanID:     0x3546,
		Suppressed: true,
	}

	carrier := map[string]string{
		"key1":        "value1",
		"x-instana-t": "0000000000001314",
		"X-INSTANA-S": "0000000000001314",
		"X-Instana-L": "1",
	}

	require.NoError(t, tracer.Inject(sc, ot.TextMap, ot.TextMapCarrier(carrier)))

	assert.Equal(t, map[string]string{
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
				"x-instana-t":     "10000000000001314",
				"X-INSTANA-S":     "2435",
				"X-Instana-L":     "1",
				"X-Instana-B-Foo": "bar",
			},
			Expected: instana.SpanContext{
				TraceIDHi: 0x1,
				TraceID:   0x1314,
				SpanID:    0x2435,
				Baggage: map[string]string{
					"Foo": "bar",
				},
			},
		},
		"tracing disabled": {
			Carrier: map[string]string{
				"Authorization": "Basic 123",
				"x-instana-t":   "10000000000001314",
				"X-INSTANA-S":   "2435",
				"X-Instana-L":   "0",
			},
			Expected: instana.SpanContext{
				TraceIDHi:  0x1,
				TraceID:    0x1314,
				SpanID:     0x2435,
				Suppressed: true,
				Baggage:    map[string]string{},
			},
		},
		"tracing disabled, no instana context": {
			Carrier: map[string]string{
				"Authorization": "Basic 123",
				"X-Instana-L":   "0",
			},
			Expected: instana.SpanContext{
				Suppressed: true,
				Baggage:    map[string]string{},
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
			defer instana.ShutdownSensor()

			sc, err := tracer.Extract(ot.TextMap, ot.TextMapCarrier(example.Carrier))
			require.NoError(t, err)

			assert.Equal(t, example.Expected, sc)
		})
	}
}

func TestTracer_Extract_TextMap_NoContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	carrier := map[string]string{
		"key": "value",
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
		"missing span id": {
			"x-instana-t": "1314",
			"x-instana-l": "1",
		},
		"malformed trace id": {
			"x-instana-t": "wrong",
			"x-instana-s": "1314",
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
			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
			defer instana.ShutdownSensor()

			_, err := tracer.Extract(ot.TextMap, ot.TextMapCarrier(carrier))
			assert.Equal(t, ot.ErrSpanContextCorrupted, err)
		})
	}
}

type textMapWithRemoveAll struct {
	ot.TextMapCarrier
}

func (c *textMapWithRemoveAll) RemoveAll() {
	for k := range c.TextMapCarrier {
		delete(c.TextMapCarrier, k)
	}
}

func TestTracer_Inject_CarrierWithRemoveAll_SuppressedTrace(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	sc := instana.SpanContext{
		TraceIDHi:  0x1,
		TraceID:    0x2435,
		SpanID:     0x3546,
		Suppressed: true,
	}

	carrier := map[string]string{
		"x-instana-t": "0000000000001314",
		"X-INSTANA-S": "0000000000001314",
		"X-Instana-L": "1",
	}

	require.NoError(t, tracer.Inject(sc, ot.TextMap, &textMapWithRemoveAll{carrier}))

	assert.Equal(t, map[string]string{
		"X-Instana-L": "0",
	}, carrier)
}
