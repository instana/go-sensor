// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/opentracing/opentracing-go"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/aws/aws-sdk-go/service/lambda"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartInvokeLambdaSpan_WithActiveSpan(t *testing.T) {
	const funcName = "test-function"
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	parentSp := sensor.Tracer().StartSpan("testing")

	req := newInvokeRequest(funcName)
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartInvokeLambdaSpan(req, sensor)

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

	assert.Equal(t, "aws.lambda.invoke", invokeSpan.Name)
	assert.Empty(t, invokeSpan.Ec)

	assert.IsType(t, instana.AWSLambdaInvokeSpanData{}, invokeSpan.Data)
	data := invokeSpan.Data.(instana.AWSLambdaInvokeSpanData)

	assert.Equal(t, instana.AWSInvokeSpanTags{
		FunctionName:   funcName,
		InvocationType: lambda.InvocationTypeRequestResponse,
	}, data.Tags)

	clientContext, err := (&instaawssdk.LambdaClientContext{
		Custom: map[string]string{
			"x-instana-l": "1",
			"x-instana-s": instana.FormatID(invokeSpan.SpanID),
			"x-instana-t": instana.FormatID(invokeSpan.TraceID),
		},
	}).Base64JSON()

	assert.NoError(t, err)
	assert.Equal(t, req.Params, &lambda.InvokeInput{
		ClientContext:  &clientContext,
		FunctionName:   aws.String(funcName),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
	})
}

func TestStartInvokeLambdaSpan_NoActiveSpan(t *testing.T) {
	const funcName = "test-function"

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)
	defer instana.ShutdownSensor()

	req := newInvokeRequest(funcName)
	instaawssdk.StartInvokeLambdaSpan(req, sensor)

	_, ok := instana.SpanFromContext(req.Context())
	require.False(t, ok)
}

func TestFinalizeInvoke_NoError(t *testing.T) {
	const funcName = "test-function"

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("aws.lambda.invoke", opentracing.Tags{
		"function": funcName,
		"type":     lambda.InvocationTypeRequestResponse,
	})

	req := newInvokeRequest(funcName)
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))

	instaawssdk.FinalizeInvokeLambdaSpan(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	invokeSpan := spans[0]

	assert.IsType(t, instana.AWSLambdaInvokeSpanData{}, invokeSpan.Data)
	data := invokeSpan.Data.(instana.AWSLambdaInvokeSpanData)

	assert.Equal(t, instana.AWSInvokeSpanTags{
		FunctionName:   funcName,
		InvocationType: lambda.InvocationTypeRequestResponse,
	}, data.Tags)

	assert.Equal(t, req.Params, &lambda.InvokeInput{
		FunctionName:   aws.String(funcName),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
	})
}

func TestFinalizeInvokeLambdaSpan_WithError(t *testing.T) {
	const funcName = "test-function"

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("aws.lambda.invoke", opentracing.Tags{
		"function": funcName,
		"type":     lambda.InvocationTypeRequestResponse,
	})

	req := newInvokeRequest(funcName)
	req.Error = awserr.New("42", "test error", errors.New("an error occurred"))
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))

	instaawssdk.FinalizeInvokeLambdaSpan(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	invokeSpan := spans[0]

	assert.IsType(t, instana.AWSLambdaInvokeSpanData{}, invokeSpan.Data)
	data := invokeSpan.Data.(instana.AWSLambdaInvokeSpanData)

	assert.Equal(t, instana.AWSInvokeSpanTags{
		FunctionName:   funcName,
		InvocationType: lambda.InvocationTypeRequestResponse,
		Error:          req.Error.Error(),
	}, data.Tags)

	assert.Equal(t, req.Params, &lambda.InvokeInput{
		FunctionName:   aws.String(funcName),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
	})
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
