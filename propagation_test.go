package instana_test

import (
	"fmt"
	"net/http"
	"testing"

	instana "github.com/instana/go-sensor"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpanPropagator(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("parent-span")
	sp.SetBaggageItem("foo", "bar")

	tmc := opentracing.HTTPHeadersCarrier(http.Header{})
	tests := []struct {
		typ, carrier interface{}
	}{
		{opentracing.HTTPHeaders, tmc},
		{opentracing.TextMap, tmc},
	}

	for i, test := range tests {
		require.NoError(t, tracer.Inject(sp.Context(), test.typ, test.carrier), "span %d", i)

		injectedContext, err := tracer.Extract(test.typ, test.carrier)
		require.NoError(t, err, "span %d", i)

		child := tracer.StartSpan("child-span", opentracing.ChildOf(injectedContext))
		child.Finish()
	}

	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, len(tests)+1)

	// The last span is the parent one
	exp, spans := spans[len(spans)-1], spans[:len(spans)-1]
	for i, span := range spans {
		assert.Equal(t, exp.TraceID, span.TraceID, "span %d", i)
		assert.Equal(t, exp.SpanID, span.ParentID, "span %d", i)
	}
}

func TestCaseSensitiveHeaderPropagation(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	// Simulate an existing root span
	metadata := map[string]string{
		"X-Instana-T":     "1314",
		"X-Instana-S":     "1314",
		"X-Instana-L":     "1",
		"X-Instana-B-Foo": "bar",
	}

	tmc1 := opentracing.TextMapCarrier(metadata)
	tmc2 := opentracing.TextMapCarrier(make(map[string]string))

	for k, v := range tmc1 {
		tmc2[k] = v
	}

	tests := []struct {
		typ, carrier interface{}
	}{
		{opentracing.HTTPHeaders, tmc1},
		{opentracing.TextMap, tmc2},
	}

	for i, test := range tests {
		// Extract the existing context
		injectedContext, err := tracer.Extract(test.typ, test.carrier)
		require.NoError(t, err, "span %d", i)

		// Start a new child span and overwrite the baggage key
		child := tracer.StartSpan("child-span", opentracing.ChildOf(injectedContext))
		child.SetBaggageItem("foo", "baz")

		// Inject the context into the metadata
		require.NoError(t, tracer.Inject(child.Context(), test.typ, test.carrier), "span %d", i)

		child.Finish()
		assert.Equal(t, child.BaggageItem("foo"), "baz")
	}

	for i, s := range recorder.GetQueuedSpans() {
		assert.EqualValues(t, 0x1314, s.ParentID, "span %d", i)
		assert.NotEqual(t, 0x1314, s.SpanID, "span %d", i)
	}

}

func TestSingleHeaderPropagation(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	// Simulate an existing root span
	metadata := http.Header{}
	metadata.Set("X-Instana-T", "1314")
	metadata.Set("X-Instana-S", "1314")
	metadata.Set("X-Instana-L", "1")
	metadata.Set("X-Instana-B-Foo", "bar")

	tmc1 := opentracing.HTTPHeadersCarrier(metadata)

	// Extract the existing context
	injectedContext, err := tracer.Extract(opentracing.HTTPHeaders, tmc1)
	require.NoError(t, err)

	// Start a new child span and overwrite the baggage key
	child := tracer.StartSpan("child-span", opentracing.ChildOf(injectedContext))
	child.SetBaggageItem("foo", "baz")

	// Inject the context into the metadata
	require.NoError(t, tracer.Inject(child.Context(), opentracing.HTTPHeaders, tmc1))

	child.Finish()

	s := recorder.GetQueuedSpans()[0]
	assert.Equal(t, child.BaggageItem("foo"), "baz")

	assert.Equal(t, []string{fmt.Sprintf("%x", s.SpanID)}, http.Header(tmc1)["X-Instana-S"])

	for _, s := range recorder.GetQueuedSpans() {
		assert.Equal(t, 0x1314, s.ParentID)
		assert.NotEqual(t, 0x1314, s.SpanID)
	}

}
