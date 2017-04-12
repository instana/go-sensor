package instana

import (
	"testing"

	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
)

func TestGetServiceName(t *testing.T) {
	opts := Options{
		TestMode: true,
		LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewRecorder(opts.TestMode)
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

func TestSpanKind(t *testing.T) {
	opts := Options{
		TestMode: true,
		LogLevel: Debug}

	InitSensor(&opts)
	recorder := NewRecorder(opts.TestMode)
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
