package instana_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestSpanPropagator(t *testing.T) {
	const op = "test"
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan(op)
	sp.SetBaggageItem("foo", "bar")

	tmc := opentracing.HTTPHeadersCarrier(http.Header{})
	tests := []struct {
		typ, carrier interface{}
	}{
		{opentracing.HTTPHeaders, tmc},
		{opentracing.TextMap, tmc},
	}

	for i, test := range tests {
		if err := tracer.Inject(sp.Context(), test.typ, test.carrier); err != nil {
			t.Fatalf("%d: %v", i, err)
		}

		injectedContext, err := tracer.Extract(test.typ, test.carrier)
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}

		child := tracer.StartSpan(
			op,
			opentracing.ChildOf(injectedContext))
		child.Finish()
	}

	sp.Finish()

	spans := recorder.GetQueuedSpans()
	if a, e := len(spans), len(tests)+1; a != e {
		t.Fatalf("expected %d spans, got %d", e, a)
	}

	// The last span is the original one.
	exp, spans := spans[len(spans)-1], spans[:len(spans)-1]
	exp.Duration = uint64(time.Duration(123))
	// exp.Timestamp = uint64(time.Time{}.Add(1))

	for i, span := range spans {
		if a, e := *span.ParentID, exp.SpanID; a != e {
			t.Fatalf("%d: ParentID %d does not match expectation %d", i, a, e)
		} else {
			// Prepare for comparison.
			span.SpanID, span.ParentID = exp.SpanID, nil
			span.Duration, span.Timestamp = exp.Duration, exp.Timestamp
		}

		if a, e := span.TraceID, exp.TraceID; a != e {
			t.Fatalf("%d: TraceID changed from %d to %d", i, e, a)
		}
	}
}

func TestCaseSensitiveHeaderPropagation(t *testing.T) {
	var (
		op                 = "test"
		spanParentIDBase64 = int64(4884)
		spanParentIDString = "1314"
	)

	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	// Simulate an existing root span
	metadata := make(map[string]string)
	metadata["X-Instana-T"] = spanParentIDString
	metadata["X-Instana-S"] = spanParentIDString
	metadata["X-Instana-L"] = "1"
	metadata["X-Instana-B-Foo"] = "bar"

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
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		// Start a new child span and overwrite the baggage key
		child := tracer.StartSpan(
			op,
			opentracing.ChildOf(injectedContext))
		child.SetBaggageItem("foo", "baz")

		// Inject the context into the metadata
		if err := tracer.Inject(child.Context(), test.typ, test.carrier); err != nil {
			t.Fatalf("%d: %v", i, err)
		}

		child.Finish()
		assert.Equal(t, child.BaggageItem("foo"), "baz")

	}

	for _, s := range recorder.GetQueuedSpans() {
		assert.Equal(t, spanParentIDBase64, *s.ParentID)
		assert.NotEqual(t, spanParentIDBase64, s.SpanID)
	}

}

func TestSingleHeaderPropagation(t *testing.T) {
	var (
		op                 = "test"
		spanParentIDBase64 = int64(4884)
		spanParentIDString = "1314"
	)

	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	// Simulate an existing root span
	metadata := make(http.Header)
	metadata.Set("X-Instana-T", spanParentIDString)
	metadata.Set("X-Instana-S", spanParentIDString)
	metadata.Set("X-Instana-L", "1")
	metadata.Set("X-Instana-B-Foo", "bar")
	tmc1 := opentracing.HTTPHeadersCarrier(metadata)

	tests := []struct {
		typ, carrier interface{}
	}{
		{opentracing.HTTPHeaders, tmc1},
	}

	for i, test := range tests {
		// Extract the existing context
		injectedContext, err := tracer.Extract(test.typ, test.carrier)
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		// Start a new child span and overwrite the baggage key
		child := tracer.StartSpan(
			op,
			opentracing.ChildOf(injectedContext))
		child.SetBaggageItem("foo", "baz")

		// Inject the context into the metadata
		if err := tracer.Inject(child.Context(), test.typ, test.carrier); err != nil {
			t.Fatalf("%d: %v", i, err)
		}

		child.Finish()
		s := recorder.GetQueuedSpans()[0]
		assert.Equal(t, child.BaggageItem("foo"), "baz")
		assert.Equal(t, []string{fmt.Sprintf("%x", s.SpanID)}, http.Header(test.carrier.(opentracing.HTTPHeadersCarrier))["X-Instana-S"])
	}

	for _, s := range recorder.GetQueuedSpans() {
		assert.Equal(t, spanParentIDBase64, *s.ParentID)
		assert.NotEqual(t, spanParentIDBase64, s.SpanID)
	}

}
