package instana

import (
	"testing"

	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
)

func TestGetServiceNameByTracer(t *testing.T) {
	opts := Options{LogLevel: Debug, Service: "tracer-named-service"}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan("http-client")
	sp.SetTag(string(ext.SpanKind), "exit")
	sp.SetTag("http.status", 200)
	sp.SetTag(string(ext.HTTPMethod), "GET")

	sp.Finish()

	rawSpan := sp.(*spanS).raw
	serviceName := getServiceName(rawSpan)
	assert.EqualValues(t, "tracer-named-service", serviceName, "Wrong Service Name")
}

func TestGetServiceNameByHTTP(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan("http-client")
	sp.SetTag(string(ext.SpanKind), "exit")
	sp.SetTag("http.status", 200)
	sp.SetTag("http.url", "https://www.instana.com/product/")
	sp.SetTag(string(ext.HTTPMethod), "GET")

	sp.Finish()

	rawSpan := sp.(*spanS).raw
	serviceName := getServiceName(rawSpan)
	assert.EqualValues(t, "https://www.instana.com/product/", serviceName, "Wrong Service Name")
}

func TestGetServiceNameByComponent(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan("http-client")
	sp.SetTag(string(ext.SpanKind), "exit")
	sp.SetTag("http.status", 200)
	sp.SetTag("component", "component-named-service")
	sp.SetTag(string(ext.HTTPMethod), "GET")

	sp.Finish()

	rawSpan := sp.(*spanS).raw
	serviceName := getServiceName(rawSpan)
	assert.EqualValues(t, "component-named-service", serviceName, "Wrong Service Name")
}

func TestSpanKind(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	// Exit
	sp := tracer.StartSpan("http-client")
	sp.SetTag(string(ext.SpanKind), "exit")
	sp.Finish()
	rawSpan := sp.(*spanS).raw
	kind := getSpanKind(rawSpan)
	assert.EqualValues(t, "exit", kind, "Wrong span kind")

	// Entry
	sp = tracer.StartSpan("http-client")
	sp.SetTag(string(ext.SpanKind), "entry")
	sp.Finish()
	rawSpan = sp.(*spanS).raw
	kind = getSpanKind(rawSpan)
	assert.EqualValues(t, "entry", kind, "Wrong span kind")

	// Consumer
	sp = tracer.StartSpan("http-client")
	sp.SetTag(string(ext.SpanKind), "consumer")
	sp.Finish()
	rawSpan = sp.(*spanS).raw
	kind = getSpanKind(rawSpan)
	assert.EqualValues(t, "entry", kind, "Wrong span kind")

	// Producer
	sp = tracer.StartSpan("http-client")
	sp.SetTag(string(ext.SpanKind), "producer")
	sp.Finish()
	rawSpan = sp.(*spanS).raw
	kind = getSpanKind(rawSpan)
	assert.EqualValues(t, "exit", kind, "Wrong span kind")
}

func TestGetTag(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	// Exit
	sp := tracer.StartSpan("http-client")
	sp.SetTag("foo", "bar")
	sp.Finish()
	rawSpan := sp.(*spanS).raw
	tag := getTag(rawSpan, "foo")
	assert.EqualValues(t, "bar", tag, "getTag unexpected return value")

	sp = tracer.StartSpan("http-client")
	sp.Finish()
	rawSpan = sp.(*spanS).raw
	tag = getTag(rawSpan, "magic")
	assert.EqualValues(t, "", tag, "getTag should return empty string for non-existent tags")
}

func TestGetIntTag(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan("http-client")
	sp.SetTag("one", 1)
	sp.SetTag("two", "twotwo")
	sp.Finish()
	rawSpan := sp.(*spanS).raw
	tag := getIntTag(rawSpan, "one")
	assert.EqualValues(t, 1, tag, "geIntTag unexpected return value")

	// Non-existent
	tag = getIntTag(rawSpan, "thirtythree")
	assert.EqualValues(t, -1, tag, "geIntTag should return -1 for non-existent tags")

	// Non-Int value (it's a string)
	tag = getIntTag(rawSpan, "two")
	assert.EqualValues(t, -1, tag, "geIntTag should return -1 for non-int tags")
}

func TestGetStringTag(t *testing.T) {
	opts := Options{LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewTestRecorder()
	tracer := NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan("http-client")
	sp.SetTag("int", 1)
	sp.SetTag("float", 2.3420493)
	sp.SetTag("two", "twotwo")
	sp.Finish()
	rawSpan := sp.(*spanS).raw
	tag := getStringTag(rawSpan, "two")
	assert.EqualValues(t, "twotwo", tag, "geStringTag unexpected return value")

	// Non-existent
	tag = getStringTag(rawSpan, "thirtythree")
	assert.EqualValues(t, "", tag, "getStringTag should return empty string for non-existent tags")

	// Non-string value (it's an int)
	tag = getStringTag(rawSpan, "int")
	assert.EqualValues(t, "1", tag, "geStringTag should return string for non-string tag values")

	// Non-string value (it's an float)
	tag = getStringTag(rawSpan, "float")
	assert.EqualValues(t, "2.3420493", tag, "geStringTag should return string for non-string tag values")
}
