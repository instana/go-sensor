package instana_test

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	"github.com/instana/golang-sensor"
	ext "github.com/opentracing/opentracing-go/ext"
)

func TestRecorderSDKReporting(t *testing.T) {
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan("http-client")
	sp.SetTag(string(ext.SpanKind), "exit")
	sp.SetTag("http.status", 200)
	sp.SetTag("http.url", "https://www.instana.com/product/")
	sp.SetTag(string(ext.HTTPMethod), "GET")

	sp.Finish()

	spans := recorder.GetSpans()
	j, _ := json.MarshalIndent(spans, "", "  ")
	log.Printf("spans:", bytes.NewBuffer(j))
}
