package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
)

func TestRecorderBasics(t *testing.T) {
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	span := tracer.StartSpan("http-client")
	span.SetTag(string(ext.SpanKind), "exit")
	span.SetTag("http.status", 200)
	span.SetTag("http.url", "https://www.instana.com/product/")
	span.SetTag(string(ext.HTTPMethod), "GET")
	span.Finish()

	// Validate GetQueuedSpans returns queued spans and clears the queue
	spans := recorder.GetQueuedSpans()
	assert.Equal(t, 1, len(spans))
	assert.Equal(t, 0, recorder.QueuedSpansCount())
}
