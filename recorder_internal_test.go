package instana

import (
	"testing"

	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
)

func TestSpan_Kind(t *testing.T) {
	tracer := NewTracerWithEverything(&Options{}, NewTestRecorder())

	// Exit
	span := tracer.StartSpan("http-client")
	span.SetTag(string(ext.SpanKind), "exit")
	assert.Equal(t, ExitSpanKind, span.(*spanS).Kind())

	// Entry
	span = tracer.StartSpan("http-client")
	span.SetTag(string(ext.SpanKind), "entry")
	assert.EqualValues(t, EntrySpanKind, span.(*spanS).Kind())

	// Consumer
	span = tracer.StartSpan("http-client")
	span.SetTag(string(ext.SpanKind), "consumer")
	assert.EqualValues(t, EntrySpanKind, span.(*spanS).Kind())

	// Producer
	span = tracer.StartSpan("http-client")
	span.SetTag(string(ext.SpanKind), "producer")
	assert.EqualValues(t, ExitSpanKind, span.(*spanS).Kind())

	// Intermediate
	span = tracer.StartSpan("http-client")
	assert.EqualValues(t, IntermediateSpanKind, span.(*spanS).Kind())
}
