package instana

import (
	"testing"

	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
)

func TestSpanKind(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	// Exit
	span := tracer.StartSpan("http-client")
	span.SetTag(string(ext.SpanKind), "exit")
	span.Finish()
	kind := span.(*spanS).getSpanKindTag()
	assert.EqualValues(t, "exit", kind, "Wrong span kind")

	// Entry
	span = tracer.StartSpan("http-client")
	span.SetTag(string(ext.SpanKind), "entry")
	span.Finish()
	kind = span.(*spanS).getSpanKindTag()
	assert.EqualValues(t, "entry", kind, "Wrong span kind")

	// Consumer
	span = tracer.StartSpan("http-client")
	span.SetTag(string(ext.SpanKind), "consumer")
	span.Finish()
	kind = span.(*spanS).getSpanKindTag()
	assert.EqualValues(t, "entry", kind, "Wrong span kind")

	// Producer
	span = tracer.StartSpan("http-client")
	span.SetTag(string(ext.SpanKind), "producer")
	span.Finish()
	kind = span.(*spanS).getSpanKindTag()
	assert.EqualValues(t, "exit", kind, "Wrong span kind")
}

func TestGetTag(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	// Exit
	span := tracer.StartSpan("http-client")
	span.SetTag("foo", "bar")
	span.Finish()
	tag := span.(*spanS).getTag("foo")
	assert.EqualValues(t, "bar", tag, "getTag unexpected return value")

	span = tracer.StartSpan("http-client")
	span.Finish()
	tag = span.(*spanS).getTag("magic")
	assert.EqualValues(t, "", tag, "getTag should return empty string for non-existent tags")
}

func TestGetIntTag(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	span := tracer.StartSpan("http-client")
	span.SetTag("one", 1)
	span.SetTag("two", "twotwo")
	span.Finish()
	tag := span.(*spanS).getIntTag("one")
	assert.EqualValues(t, 1, tag, "geIntTag unexpected return value")

	// Non-existent
	tag = span.(*spanS).getIntTag("thirtythree")
	assert.EqualValues(t, -1, tag, "geIntTag should return -1 for non-existent tags")

	// Non-Int value (it's a string)
	tag = span.(*spanS).getIntTag("two")
	assert.EqualValues(t, -1, tag, "geIntTag should return -1 for non-int tags")
}

func TestGetStringTag(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	span := tracer.StartSpan("http-client")
	span.SetTag("int", 1)
	span.SetTag("float", 2.3420493)
	span.SetTag("two", "twotwo")
	span.Finish()
	tag := span.(*spanS).getStringTag("two")
	assert.EqualValues(t, "twotwo", tag, "geStringTag unexpected return value")

	// Non-existent
	tag = span.(*spanS).getStringTag("thirtythree")
	assert.EqualValues(t, "", tag, "getStringTag should return empty string for non-existent tags")

	// Non-string value (it's an int)
	tag = span.(*spanS).getStringTag("int")
	assert.EqualValues(t, "1", tag, "geStringTag should return string for non-string tag values")

	// Non-string value (it's an float)
	tag = span.(*spanS).getStringTag("float")
	assert.EqualValues(t, "2.3420493", tag, "geStringTag should return string for non-string tag values")
}

func TestGetHostName(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	span := tracer.StartSpan("http-client")
	span.SetTag("int", 1)
	span.SetTag("float", 2.3420493)
	span.SetTag("two", "twotwo")
	span.Finish()
	hostname := span.(*spanS).getHostName()
	assert.True(t, len(hostname) > 0, "must return a valid string value")
}
