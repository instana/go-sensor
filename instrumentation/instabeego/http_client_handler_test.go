// (c) Copyright IBM Corp. 2024

//go:build go1.18
// +build go1.18

package instabeego_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/client/httplib"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instabeego"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstrumentRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "beego-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	sp := c.StartSpan("client-call")
	sp.SetTag(string(ext.SpanKind), "entry")

	ctx := instana.ContextWithSpan(context.Background(), sp)

	req := httplib.NewBeegoRequestWithCtx(ctx, server.URL, http.MethodGet)
	instabeego.InstrumentRequest(c, req)

	response, err := req.Response()
	require.NoError(t, err)

	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	exitSpan := spans[0]
	spanData := exitSpan.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Method:   response.Request.Method,
		Status:   http.StatusOK,
		Path:     "",
		URL:      response.Request.URL.String(),
		Host:     "",
		Protocol: "",
		Params:   "",
		Headers:  nil,
	}, spanData.Tags)
}
