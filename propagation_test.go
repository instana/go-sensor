package instana_test

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/instana/golang-sensor"
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

	spans := recorder.GetSpans()
	if a, e := len(spans), len(tests)+1; a != e {
		t.Fatalf("expected %d spans, got %d", e, a)
	}

	// The last span is the original one.
	exp, spans := spans[len(spans)-1], spans[:len(spans)-1]
	exp.Duration = uint64(time.Duration(123))
	exp.Raw.Start = time.Time{}.Add(1)

	for i, sp := range spans {
		if a, e := sp.Raw.ParentSpanID, exp.Raw.Context.SpanID; a != e {
			t.Fatalf("%d: ParentSpanID %d does not match expectation %d", i, a, e)
		} else {
			// Prepare for comparison.
			sp.Raw.Context.SpanID, sp.Raw.ParentSpanID = exp.Raw.Context.SpanID, 0
			sp.Duration, sp.Raw.Start = exp.Duration, exp.Raw.Start
		}

		if a, e := sp.Raw.Context.TraceID, exp.Raw.Context.TraceID; a != e {
			t.Fatalf("%d: TraceID changed from %d to %d", i, e, a)
		}

		if !reflect.DeepEqual(exp, sp) {
			t.Fatalf("%d: wanted %+v, got %+v", i, spew.Sdump(exp), spew.Sdump(sp))
		}
	}
}

func TestCaseSensitiveHeaderPropagation(t *testing.T) {
	var (
		op                 = "test"
		spanParentIdBase64 = uint64(4884)
		spanParentIdString = "1314"
	)

	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	// Simulate an existing root span
	metadata := make(map[string]string)
	metadata["X-Instana-T"] = spanParentIdString
	metadata["X-Instana-S"] = spanParentIdString
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

	for _, s := range recorder.GetSpans() {
		assert.Equal(t, spanParentIdBase64, s.Raw.ParentSpanID)
		assert.NotEqual(t, spanParentIdBase64, s.Raw.Context.SpanID)
	}

}
