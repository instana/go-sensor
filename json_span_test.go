package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
)

func TestRegisteredSpanType_ExtractData(t *testing.T) {
	examples := map[string]struct {
		Operation string
		Expected  interface{}
	}{
		"net/http.Server": {
			Operation: "g.http",
			Expected:  instana.HTTPSpanData{},
		},
		"net/http.Client": {
			Operation: "http",
			Expected:  instana.HTTPSpanData{},
		},
		"golang.google.org/gppc.Server": {
			Operation: "rpc-server",
			Expected:  instana.RPCSpanData{},
		},
		"sdk": {
			Operation: "test",
			Expected:  instana.SDKSpanData{},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

			sp := tracer.StartSpan(example.Operation)
			sp.Finish()

			spans := recorder.GetQueuedSpans()
			assert.Equal(t, 1, len(spans))
			span := spans[0]

			assert.IsType(t, example.Expected, span.Data)
		})
	}
}
