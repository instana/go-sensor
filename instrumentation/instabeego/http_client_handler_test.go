// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instabeego

import (
	"context"
	"net/http"
	"testing"

	"github.com/beego/beego/v2/client/httplib"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstrumentRequest(t *testing.T) {

	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(nil, recorder)
	defer instana.ShutdownSensor()
	sensor := instana.NewSensorWithTracer(tracer)

	sp := sensor.StartSpan("client-call")
	sp.SetTag(string(ext.SpanKind), "entry")

	defer sp.Finish()

	ctx := instana.ContextWithSpan(context.Background(), sp)

	req := httplib.NewBeegoRequestWithCtx(ctx, "https://www.instana.com", http.MethodGet)
	InstrumentRequest(sensor, req)

	response, err := req.Response()
	require.NoError(t, err)
	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 3)
	latestSpan := spans[2]
	spanData := latestSpan.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Method:   response.Request.Method,
		Status:   http.StatusOK,
		Path:     "",
		URL:      response.Request.URL.String(),
		Host:     response.Request.Host,
		Protocol: "",
		Params:   "",
		Headers:  nil,
	}, spanData.Tags)
}
