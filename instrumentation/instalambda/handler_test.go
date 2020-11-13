package instalambda_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instalambda"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

func TestTraceHandlerFunc_APIGatewayEvent(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(instana.DefaultOptions(), recorder))

	payload, err := ioutil.ReadFile("testdata/apigw_event.json")
	require.NoError(t, err)

	h := instalambda.TraceHandlerFunc(func(ctx context.Context, evt *events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		_, ok := instana.SpanFromContext(ctx)
		assert.True(t, ok)

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       "OK",
		}, nil
	}, sensor)

	lambdacontext.FunctionName = "test-function"
	lambdacontext.FunctionVersion = "42"

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID:       "req1",
		InvokedFunctionArn: "aws:test-function",
	})

	resp, err := h.Invoke(ctx, payload)
	require.NoError(t, err)

	assert.JSONEq(t, `{"statusCode":200,"headers":null,"multiValueHeaders":null,"body":"OK"}`, string(resp))

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	require.Equal(t, "aws.lambda.entry", span.Name)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

	assert.Equal(t, instana.AWSLambdaSpanData{
		Snapshot: instana.AWSLambdaSpanTags{
			ARN:     "aws:test-function:42",
			Runtime: "go",
			Name:    "test-function",
			Version: "42",
			Trigger: "aws:api.gateway",
		},
		HTTP: &instana.HTTPSpanTags{
			URL:          "/",
			Method:       "GET",
			PathTemplate: "/",
			Params:       "multisecret=key1&multisecret=key2&q=term&secret=key&value=1&value=2",
		},
	}, span.Data)
}

func TestTraceHandlerFunc_ALBEvent(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(instana.DefaultOptions(), recorder))

	payload, err := ioutil.ReadFile("testdata/alb_event.json")
	require.NoError(t, err)

	h := instalambda.TraceHandlerFunc(func(ctx context.Context, evt *events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
		_, ok := instana.SpanFromContext(ctx)
		assert.True(t, ok)

		return events.ALBTargetGroupResponse{
			StatusCode: http.StatusOK,
			Body:       "OK",
		}, nil
	}, sensor)

	lambdacontext.FunctionName = "test-function"
	lambdacontext.FunctionVersion = "42"

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID:       "req1",
		InvokedFunctionArn: "aws:test-function",
	})

	resp, err := h.Invoke(ctx, payload)
	require.NoError(t, err)

	assert.JSONEq(t, `{"statusCode":200,"statusDescription":"","headers":null,"multiValueHeaders":null,"body":"OK","isBase64Encoded":false}`, string(resp))

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	require.Equal(t, "aws.lambda.entry", span.Name)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

	assert.Equal(t, instana.AWSLambdaSpanData{
		Snapshot: instana.AWSLambdaSpanTags{
			ARN:     "aws:test-function:42",
			Runtime: "go",
			Name:    "test-function",
			Version: "42",
			Trigger: "aws:application.load.balancer",
		},
		HTTP: &instana.HTTPSpanTags{
			URL:    "/lambda",
			Method: "GET",
			Params: "query=1234ABCD",
		},
	}, span.Data)
}

func TestTraceHandlerFunc_CloudWatchEvent(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(instana.DefaultOptions(), recorder))

	payload, err := ioutil.ReadFile("testdata/cw_event.json")
	require.NoError(t, err)

	h := instalambda.TraceHandlerFunc(func(ctx context.Context, evt *events.CloudWatchEvent) error {
		_, ok := instana.SpanFromContext(ctx)
		assert.True(t, ok)

		return nil
	}, sensor)

	lambdacontext.FunctionName = "test-function"
	lambdacontext.FunctionVersion = "42"

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID:       "req1",
		InvokedFunctionArn: "aws:test-function",
	})

	_, err = h.Invoke(ctx, payload)
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	require.Equal(t, "aws.lambda.entry", span.Name)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

	assert.Equal(t, instana.AWSLambdaSpanData{
		Snapshot: instana.AWSLambdaSpanTags{
			ARN:     "aws:test-function:42",
			Runtime: "go",
			Name:    "test-function",
			Version: "42",
			Trigger: "aws:cloudwatch.events",
			CloudWatch: &instana.AWSLambdaCloudWatchSpanTags{
				Events: &instana.AWSLambdaCloudWatchEventTags{
					ID:        "cdc73f9d-aea9-11e3-9d5a-835b769c0d9c",
					Resources: []string{"arn:aws:events:us-east-1:123456789012:rule/my-schedule"},
				},
			},
		},
	}, span.Data)
}

func TestTraceHandlerFunc_CloudWatchLogsEvent(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(instana.DefaultOptions(), recorder))

	payload, err := ioutil.ReadFile("testdata/cw_logs_event.json")
	require.NoError(t, err)

	h := instalambda.TraceHandlerFunc(func(ctx context.Context, evt *events.CloudwatchLogsEvent) error {
		_, ok := instana.SpanFromContext(ctx)
		assert.True(t, ok)

		return nil
	}, sensor)

	lambdacontext.FunctionName = "test-function"
	lambdacontext.FunctionVersion = "42"

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID:       "req1",
		InvokedFunctionArn: "aws:test-function",
	})

	_, err = h.Invoke(ctx, payload)
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	require.Equal(t, "aws.lambda.entry", span.Name)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

	assert.Equal(t, instana.AWSLambdaSpanData{
		Snapshot: instana.AWSLambdaSpanTags{
			ARN:     "aws:test-function:42",
			Runtime: "go",
			Name:    "test-function",
			Version: "42",
			Trigger: "aws:cloudwatch.logs",
			CloudWatch: &instana.AWSLambdaCloudWatchSpanTags{
				Logs: &instana.AWSLambdaCloudWatchLogsTags{
					Group:  "testLogGroup",
					Stream: "testLogStream",
					Events: []string{
						"[ERROR] First test message",
						"[ERROR] Second test message",
					},
				},
			},
		},
	}, span.Data)
}

func TestTraceHandlerFunc_CloudWatchLogsEvent_DecodeError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(instana.DefaultOptions(), recorder))

	payload, err := ioutil.ReadFile("testdata/cw_logs_broken_event.json")
	require.NoError(t, err)

	h := instalambda.TraceHandlerFunc(func(ctx context.Context, evt *events.CloudwatchLogsEvent) error {
		_, ok := instana.SpanFromContext(ctx)
		assert.True(t, ok)

		return nil
	}, sensor)

	lambdacontext.FunctionName = "test-function"
	lambdacontext.FunctionVersion = "42"

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID:       "req1",
		InvokedFunctionArn: "aws:test-function",
	})

	_, err = h.Invoke(ctx, payload)
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	require.Equal(t, "aws.lambda.entry", span.Name)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

	assert.Equal(t, instana.AWSLambdaSpanData{
		Snapshot: instana.AWSLambdaSpanTags{
			ARN:     "aws:test-function:42",
			Runtime: "go",
			Name:    "test-function",
			Version: "42",
			Trigger: "aws:cloudwatch.logs",
			CloudWatch: &instana.AWSLambdaCloudWatchSpanTags{
				Logs: &instana.AWSLambdaCloudWatchLogsTags{
					DecodingError: "unexpected EOF",
				},
			},
		},
	}, span.Data)
}

func TestTraceHandlerFunc_S3Event(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(instana.DefaultOptions(), recorder))

	payload, err := ioutil.ReadFile("testdata/s3_event.json")
	require.NoError(t, err)

	h := instalambda.TraceHandlerFunc(func(ctx context.Context, evt *events.S3Event) error {
		_, ok := instana.SpanFromContext(ctx)
		assert.True(t, ok)

		return nil
	}, sensor)

	lambdacontext.FunctionName = "test-function"
	lambdacontext.FunctionVersion = "42"

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID:       "req1",
		InvokedFunctionArn: "aws:test-function",
	})

	_, err = h.Invoke(ctx, payload)
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	require.Equal(t, "aws.lambda.entry", span.Name)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

	assert.Equal(t, instana.AWSLambdaSpanData{
		Snapshot: instana.AWSLambdaSpanTags{
			ARN:     "aws:test-function:42",
			Runtime: "go",
			Name:    "test-function",
			Version: "42",
			Trigger: "aws:s3",
			S3: &instana.AWSLambdaS3SpanTags{
				Events: []instana.AWSS3EventTags{
					{
						Name:   "ObjectCreated:Put",
						Bucket: "lambda-artifacts-deafc19498e3f2df",
						Object: "b21b84d653bb07b05b1e6b33684dc11b",
					},
				},
			},
		},
	}, span.Data)
}
