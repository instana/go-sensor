package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
)

func TestRegisteredSpanType(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, 1, len(spans))
	span := spans[0]

	assert.IsType(t, instana.SDKSpanData{}, span.Data)
}
