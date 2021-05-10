package instaawssdk_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/opentracing/opentracing-go"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/aws/aws-sdk-go/service/lambda"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

func TestStartInvokeSpan_WithActiveSpan(t *testing.T) {
	const funcName = "test-function"
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)

	parentSp := sensor.Tracer().StartSpan("testing")

	req := newInvokeRequest(funcName)
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartInvokeSpan(req, sensor)

	sp, ok := instana.SpanFromContext(req.Context())
	require.True(t, ok)

	sp.Finish()
	parentSp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	invokeSpan, testingSpan := spans[0], spans[1]

	assert.Equal(t, testingSpan.TraceID, invokeSpan.TraceID)
	assert.Equal(t, testingSpan.SpanID, invokeSpan.ParentID)
	assert.NotEqual(t, testingSpan.SpanID, invokeSpan.SpanID)
	assert.NotEmpty(t, invokeSpan.SpanID)

	assert.Equal(t, "invoke", invokeSpan.Name)
	assert.Empty(t, invokeSpan.Ec)

	assert.IsType(t, instana.AWSInvokeSpanData{}, invokeSpan.Data)
	data := invokeSpan.Data.(instana.AWSInvokeSpanData)

	assert.Equal(t, instana.AWSInvokeSpanTags{
		FunctionName:   funcName,
		InvocationType: lambda.InvocationTypeRequestResponse,
	}, data.Tags)
}

func TestStartInvokeSpan_NoActiveSpan(t *testing.T) {
	const funcName = "test-function"

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)

	req := newInvokeRequest(funcName)
	instaawssdk.StartInvokeSpan(req, sensor)

	_, ok := instana.SpanFromContext(req.Context())
	require.False(t, ok)
}

func TestFinalizeInvoke_NoError(t *testing.T) {
	const funcName = "test-function"

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)

	sp := sensor.Tracer().StartSpan("invoke", opentracing.Tags{
		"invoke.function": funcName,
		"invoke.type":     lambda.InvocationTypeRequestResponse,
	})

	req := newInvokeRequest(funcName)
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))

	instaawssdk.FinalizeInvokeSpan(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	invokeSpan := spans[0]

	assert.IsType(t, instana.AWSInvokeSpanData{}, invokeSpan.Data)
	data := invokeSpan.Data.(instana.AWSInvokeSpanData)

	assert.Equal(t, instana.AWSInvokeSpanTags{
		FunctionName:   funcName,
		InvocationType: lambda.InvocationTypeRequestResponse,
	}, data.Tags)
}

func TestFinalizeInvokeSpan_WithError(t *testing.T) {
	const funcName = "test-function"

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)

	sp := sensor.Tracer().StartSpan("invoke", opentracing.Tags{
		"invoke.function": funcName,
		"invoke.type":     lambda.InvocationTypeRequestResponse,
	})

	req := newInvokeRequest(funcName)
	req.Error = awserr.New("42", "test error", errors.New("an error occurred"))
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))

	instaawssdk.FinalizeInvokeSpan(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	invokeSpan := spans[0]

	assert.IsType(t, instana.AWSInvokeSpanData{}, invokeSpan.Data)
	data := invokeSpan.Data.(instana.AWSInvokeSpanData)

	assert.Equal(t, instana.AWSInvokeSpanTags{
		FunctionName:   funcName,
		InvocationType: lambda.InvocationTypeRequestResponse,
		Error:          req.Error.Error(),
	}, data.Tags)
}

func newInvokeRequest(funcName string) *request.Request {
	svc := lambda.New(unit.Session)

	op := &request.Operation{
		Name:       "Invoke",
		HTTPMethod: "POST",
		HTTPPath:   "/2015-03-31/functions/{FunctionName}/invocations",
	}

	invokeType := lambda.InvocationTypeRequestResponse
	return svc.NewRequest(op, &lambda.InvokeInput{
		FunctionName:   &funcName,
		InvocationType: &invokeType,
	}, nil)
}
