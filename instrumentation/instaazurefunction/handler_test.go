// (c) Copyright IBM Corp. 2022
// (c) Copyright Instana Inc. 2022

//go:build go1.13
// +build go1.13

package instaazurefunction_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-sensor/instrumentation/instaazurefunction"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpTrigger(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instaazurefunction.WrapFunctionHandler(sensor, func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintln(writer, "Ok")
	})

	//sp := sensor.Tracer().StartSpan("testing")
	//sp.Finish()

	bodyReader := strings.NewReader(`{"Metadata":{"Headers":{"User-Agent":"curl/7.79.1"},"sys":{"MethodName":"roboshop"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/roboshop", bodyReader)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	spans := recorder.GetQueuedSpans()

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Ok\n", rec.Body.String())
	assert.Equal(t, 1, len(spans))

	azSpan := spans[0]
	data := azSpan.Data.(instana.AZFSpanData)

	assert.Equal(t, "roboshop", data.Tags.MethodName)
	assert.Equal(t, "custom", data.Tags.Runtime)
	assert.Equal(t, "httpTrigger", data.Tags.Trigger)
	assert.Equal(t, "", data.Tags.Name)
	assert.Equal(t, "", data.Tags.FunctionName)
}

func TestMultiTriggers(t *testing.T) {
	testcases := map[string]struct {
		TargetUrl string
		Payload   interface{}
		Expected  instana.AZFSpanTags
	}{
		"httpTrigger": {
			TargetUrl: "/roboshop",
			Payload:   `{"Metadata":{"Headers":{"User-Agent":"curl/7.79.1"},"sys":{"MethodName":"roboshop"}}}`,
			Expected: instana.AZFSpanTags{
				MethodName: "roboshop",
				Trigger:    "httpTrigger",
				Runtime:    "custom",
			},
		},
		"queueTrigger": {
			TargetUrl: "/roboshop",
			Payload:   `{"Metadata":{"PopReceipt":"MTROb3YyMDIyMTE6MTM6MjJiOWU4","sys":{"MethodName":"roboshop"}}}`,
			Expected: instana.AZFSpanTags{
				MethodName: "roboshop",
				Trigger:    "queueTrigger",
				Runtime:    "custom",
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

			bodyReader := strings.NewReader(testcase.Payload.(string))
			req := httptest.NewRequest(http.MethodPost, testcase.TargetUrl, bodyReader)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			spans := recorder.GetQueuedSpans()

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "Ok\n", rec.Body.String())
			assert.Equal(t, 1, len(spans))

			azSpan := spans[0]
			data := azSpan.Data.(instana.AZFSpanData)

			assert.Equal(t, testcase.Expected.MethodName, data.Tags.MethodName)
			assert.Equal(t, testcase.Expected.Runtime, data.Tags.Runtime)
			assert.Equal(t, testcase.Expected.Trigger, data.Tags.Trigger)
			assert.Equal(t, "", data.Tags.Name)
			assert.Equal(t, "", data.Tags.FunctionName)
		})
	}
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool {
	return true
}

func (alwaysReadyClient) SendSpans(spans []instana.Span) error {
	// Returning an error will result in placing the spans back in the queue of the recorder.
	// Those can be used for asserting errors.
	return errors.New("this is a mock agent clients. Cannot send data")
}

func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
