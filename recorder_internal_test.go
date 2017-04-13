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

	sp := tracer.StartSpan("http-client")
	sp.SetTag(string(ext.SpanKind), "exit")
	sp.SetTag("http.status", 200)
	sp.SetTag("http.url", "https://www.instana.com/product/")
	sp.SetTag(string(ext.HTTPMethod), "GET")

	sp.Finish()

	rawSpan := sp.(*spanS).raw
	kind := getSpanKind(rawSpan)
	assert.EqualValues(t, "exit", kind, "Wrong span kind")
}
