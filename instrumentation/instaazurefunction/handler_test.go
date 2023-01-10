// (c) Copyright IBM Corp. 2023

package instaazurefunction_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instaazurefunction"
	"github.com/stretchr/testify/assert"
)

func TestHttpTrigger(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instaazurefunction.WrapFunctionHandler(sensor, func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintln(writer, "Ok")
	})

	bodyReader := strings.NewReader(`{"Metadata":{"Headers":{"User-Agent":"curl/7.79.1"},"sys":{"MethodName":"roboshop"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/roboshop", bodyReader)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	spans := recorder.GetQueuedSpans()

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Ok\n", rec.Body.String())
	assert.Len(t, spans, 1)

	azSpan := spans[0]
	data := azSpan.Data.(instana.AZFSpanData)

	assert.Equal(t, "roboshop", data.Tags.FunctionName)
	assert.Equal(t, "custom", data.Tags.Runtime)
	assert.Equal(t, "HTTP", data.Tags.Trigger)
	assert.Equal(t, "", data.Tags.Name)
	assert.Equal(t, "", data.Tags.MethodName)
}

func TestMultiTriggers(t *testing.T) {
	testcases := map[string]struct {
		TargetUrl string
		Payload   string
		Expected  instana.AZFSpanTags
	}{
		"httpTrigger": {
			TargetUrl: "/roboshop",
			Payload:   `{"Metadata":{"Headers":{"User-Agent":"curl/7.79.1"},"sys":{"MethodName":"roboshop"}}}`,
			Expected: instana.AZFSpanTags{
				FunctionName: "roboshop",
				Trigger:      "HTTP",
				Runtime:      "custom",
			},
		},
		"queueTrigger": {
			TargetUrl: "/roboshop",
			Payload:   `{"Metadata":{"InsertionTime": "2019-10-09T17:58:31+00:00","PopReceipt":"MTROb3YyMDIyMTE6MTM6MjJiOWU4","sys":{"MethodName":"roboshop"}}}`,
			Expected: instana.AZFSpanTags{
				FunctionName: "roboshop",
				Trigger:      "Queue",
				Runtime:      "custom",
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

			h := instaazurefunction.WrapFunctionHandler(sensor, func(writer http.ResponseWriter, request *http.Request) {
				_, _ = fmt.Fprintln(writer, "Ok")
			})

			bodyReader := strings.NewReader(testcase.Payload)
			req := httptest.NewRequest(http.MethodPost, testcase.TargetUrl, bodyReader)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			spans := recorder.GetQueuedSpans()

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "Ok\n", rec.Body.String())
			assert.Len(t, spans, 1)

			azSpan := spans[0]
			data := azSpan.Data.(instana.AZFSpanData)

			assert.Equal(t, testcase.Expected.FunctionName, data.Tags.FunctionName)
			assert.Equal(t, testcase.Expected.Runtime, data.Tags.Runtime)
			assert.Equal(t, testcase.Expected.Trigger, data.Tags.Trigger)
			assert.Equal(t, "", data.Tags.Name)
			assert.Equal(t, "", data.Tags.MethodName)
		})
	}
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool {
	return true
}

func (alwaysReadyClient) SendSpans([]instana.Span) error {
	// Returning an error will result in placing the spans back in the queue of the recorder.
	// Those can be used for asserting errors.
	return errors.New("this is a mock agent clients. Cannot send data")
}

func (alwaysReadyClient) SendMetrics(acceptor.Metrics) error       { return nil }
func (alwaysReadyClient) SendEvent(*instana.EventData) error       { return nil }
func (alwaysReadyClient) SendProfiles([]autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error              { return nil }
