// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instalambda_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instalambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getOptions() *instana.Options {
	return &instana.Options{
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"X-Custom-Header-1", "X-Custom-Header-2"},
			Secrets:                instana.DefaultSecretsMatcher(),
		},
		AgentClient: alwaysReadyClient{},
	}
}

func TestNewHandler_APIGatewayEvent(t *testing.T) {
	testCases := map[string]string{
		"API_GW_Event":              "testdata/apigw_event.json",
		"API_GW_EventWithW3Context": "testdata/apigw_event_with_w3context.json",
	}

	for tc, fileName := range testCases {
		t.Run(tc, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
			defer instana.ShutdownSensor()

			payload, err := ioutil.ReadFile(fileName)
			require.NoError(t, err)

			h := instalambda.NewHandler(func(ctx context.Context, evt *events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

			assert.EqualValues(t, 0x1234, span.TraceID)
			assert.EqualValues(t, 0x4567, span.ParentID)
			assert.NotEqual(t, span.ParentID, span.SpanID)

			require.Equal(t, "aws.lambda.entry", span.Name)
			assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

			assert.Equal(t, instana.AWSLambdaSpanData{
				Snapshot: instana.AWSLambdaSpanTags{
					ARN:       "aws:test-function:42",
					Runtime:   "go",
					Name:      "test-function",
					Version:   "42",
					Trigger:   "aws:api.gateway",
					ColdStart: true,
				},
				HTTP: &instana.HTTPSpanTags{
					URL:          "/",
					Method:       "GET",
					PathTemplate: "/",
					Params:       "multisecret=%3Credacted%3E&multisecret=%3Credacted%3E&q=term&secret=%3Credacted%3E&value=1&value=2",
					Headers: map[string]string{
						"X-Custom-Header-1": "value1",
						"X-Custom-Header-2": "value2",
					},
				},
			}, span.Data)
		})
	}
}

func TestNewHandler_APIGatewayV2Event_WithW3Context(t *testing.T) {
	testCases := map[string]string{
		"API_GW_V2_Event":              "testdata/apigw_v2_event.json",
		"API_GW_V2_EventWithW3Context": "testdata/apigw_v2_event_with_w3context.json",
	}

	for tc, fileName := range testCases {
		t.Run(tc, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
			defer instana.ShutdownSensor()

			payload, err := ioutil.ReadFile(fileName)
			require.NoError(t, err)

			h := instalambda.NewHandler(func(ctx context.Context, evt *events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				_, ok := instana.SpanFromContext(ctx)
				assert.True(t, ok)

				return events.APIGatewayV2HTTPResponse{
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

			assert.JSONEq(t, `{"statusCode":200,"headers":null,"multiValueHeaders":null,"body":"OK","cookies":null}`, string(resp))

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			span := spans[0]

			assert.EqualValues(t, 0x1234, span.TraceID)
			assert.EqualValues(t, 0x4567, span.ParentID)
			assert.NotEqual(t, span.ParentID, span.SpanID)

			require.Equal(t, "aws.lambda.entry", span.Name)
			assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

			assert.Equal(t, instana.AWSLambdaSpanData{
				Snapshot: instana.AWSLambdaSpanTags{
					ARN:       "aws:test-function:42",
					Runtime:   "go",
					Name:      "test-function",
					Version:   "42",
					Trigger:   "aws:api.gateway",
					ColdStart: true,
				},
				HTTP: &instana.HTTPSpanTags{
					URL:          "/my/path",
					Method:       "POST",
					PathTemplate: "/my/{resource}",
					Params:       "q=term&secret=%3Credacted%3E",
					Headers: map[string]string{
						"X-Custom-Header-1": "value1",
						"X-Custom-Header-2": "value2",
					},
				},
			}, span.Data)
		})
	}
}

func TestNewHandler_ALBEvent(t *testing.T) {
	testCases := map[string]string{
		"ALB_Event":               "testdata/alb_event.json",
		"ALB_Event_WithW3Context": "testdata/alb_event_with_w3context.json",
	}

	for tc, fileName := range testCases {
		t.Run(tc, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
			defer instana.ShutdownSensor()

			payload, err := ioutil.ReadFile(fileName)
			require.NoError(t, err)

			h := instalambda.NewHandler(func(ctx context.Context, evt *events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
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

			assert.EqualValues(t, 0x1234, span.TraceID)
			assert.EqualValues(t, 0x4567, span.ParentID)
			assert.NotEqual(t, span.ParentID, span.SpanID)

			require.Equal(t, "aws.lambda.entry", span.Name)
			assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

			assert.Equal(t, instana.AWSLambdaSpanData{
				Snapshot: instana.AWSLambdaSpanTags{
					ARN:       "aws:test-function:42",
					Runtime:   "go",
					Name:      "test-function",
					Version:   "42",
					Trigger:   "aws:application.load.balancer",
					ColdStart: true,
				},
				HTTP: &instana.HTTPSpanTags{
					URL:    "/lambda",
					Method: "GET",
					Params: "multikey=%3Credacted%3E&multikey=%3Credacted%3E&multisecret=%3Credacted%3E&multisecret=%3Credacted%3E&query=1234ABCD&secret=%3Credacted%3E",
					Headers: map[string]string{
						"X-Custom-Header-1": "value1",
						"X-Custom-Header-2": "value2",
					},
				},
			}, span.Data)
		})
	}
}

func TestNewHandler_CloudWatchEvent(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
	defer instana.ShutdownSensor()

	payload, err := ioutil.ReadFile("testdata/cw_event.json")
	require.NoError(t, err)

	h := instalambda.NewHandler(func(ctx context.Context, evt *events.CloudWatchEvent) error {
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
			ARN:       "aws:test-function:42",
			Runtime:   "go",
			Name:      "test-function",
			Version:   "42",
			Trigger:   "aws:cloudwatch.events",
			ColdStart: true,
			CloudWatch: &instana.AWSLambdaCloudWatchSpanTags{
				Events: &instana.AWSLambdaCloudWatchEventTags{
					ID:        "cdc73f9d-aea9-11e3-9d5a-835b769c0d9c",
					Resources: []string{"arn:aws:events:us-east-1:123456789012:rule/my-schedule"},
				},
			},
		},
	}, span.Data)
}

func TestNewHandler_CloudWatchLogsEvent(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
	defer instana.ShutdownSensor()

	payload, err := ioutil.ReadFile("testdata/cw_logs_event.json")
	require.NoError(t, err)

	h := instalambda.NewHandler(func(ctx context.Context, evt *events.CloudwatchLogsEvent) error {
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
			ARN:       "aws:test-function:42",
			Runtime:   "go",
			Name:      "test-function",
			Version:   "42",
			Trigger:   "aws:cloudwatch.logs",
			ColdStart: true,
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

func TestNewHandler_CloudWatchLogsEvent_DecodeError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
	defer instana.ShutdownSensor()

	payload, err := ioutil.ReadFile("testdata/cw_logs_broken_event.json")
	require.NoError(t, err)

	h := instalambda.NewHandler(func(ctx context.Context, evt *events.CloudwatchLogsEvent) error {
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
			ARN:       "aws:test-function:42",
			Runtime:   "go",
			Name:      "test-function",
			Version:   "42",
			Trigger:   "aws:cloudwatch.logs",
			ColdStart: true,
			CloudWatch: &instana.AWSLambdaCloudWatchSpanTags{
				Logs: &instana.AWSLambdaCloudWatchLogsTags{
					DecodingError: "unexpected EOF",
				},
			},
		},
	}, span.Data)
}

func TestNewHandler_S3Event(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
	defer instana.ShutdownSensor()

	payload, err := ioutil.ReadFile("testdata/s3_event.json")
	require.NoError(t, err)

	h := instalambda.NewHandler(func(ctx context.Context, evt *events.S3Event) error {
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
			ARN:       "aws:test-function:42",
			Runtime:   "go",
			Name:      "test-function",
			Version:   "42",
			Trigger:   "aws:s3",
			ColdStart: true,
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

func TestNewHandler_SQSEvent(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
	defer instana.ShutdownSensor()

	payload, err := ioutil.ReadFile("testdata/sqs_event.json")
	require.NoError(t, err)

	h := instalambda.NewHandler(func(ctx context.Context, evt *events.SQSEvent) error {
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
			ARN:       "aws:test-function:42",
			Runtime:   "go",
			Name:      "test-function",
			Version:   "42",
			Trigger:   "aws:sqs",
			ColdStart: true,
			SQS: &instana.AWSLambdaSQSSpanTags{
				Messages: []instana.AWSSQSMessageTags{
					{Queue: "arn:aws:sqs:us-east-2:123456789012:my-queue"},
					{Queue: "arn:aws:sqs:us-east-2:123456789012:my-queue"},
				},
			},
		},
	}, span.Data)
}

func TestNewHandler_PreferInstanaHeadersToW3ContextHeaders(t *testing.T) {
	testCases := map[string]string{
		"API_GW_Event":    "testdata/apigw_v2_event_with_instana_headers_and_w3context.json",
		"API_GW_V2_Event": "testdata/apigw_event_with_instana_headers_and_w3context.json",
		"ALBEvent":        "testdata/alb_event_with_instana_headers_and_w3context.json",
	}

	for tc, fileName := range testCases {
		t.Run(tc, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
			defer instana.ShutdownSensor()

			payload, err := ioutil.ReadFile(fileName)
			require.NoError(t, err)

			h := instalambda.NewHandler(func(ctx context.Context, evt *events.APIGatewayV2HTTPRequest) {
				_, ok := instana.SpanFromContext(ctx)
				assert.True(t, ok)
			}, sensor)

			_, err = h.Invoke(lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{}), payload)
			require.NoError(t, err)

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			span := spans[0]

			assert.EqualValues(t, 0x1234, span.TraceID)
			assert.EqualValues(t, 0x4567, span.ParentID)
			assert.NotEqual(t, span.ParentID, span.SpanID)
		})
	}
}

func TestNewHandler_InvokeLambda_Success(t *testing.T) {
	testCases := map[string]*lambdacontext.LambdaContext{
		"Invoke_WithInstanaHeadersOnly": {
			AwsRequestID:       "req1",
			InvokedFunctionArn: "aws:test-function",
			ClientContext: lambdacontext.ClientContext{
				Client: lambdacontext.ClientApplication{},
				Env:    nil,
				Custom: map[string]string{
					instana.FieldT: "0000000000001234",
					instana.FieldS: "0000000000004567",
					instana.FieldL: "1",
				},
			},
		},
	}

	for tc, lc := range testCases {
		t.Run(tc, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
			defer instana.ShutdownSensor()

			h := instalambda.NewHandler(func(ctx context.Context, evt interface{}) error {
				_, ok := instana.SpanFromContext(ctx)
				assert.True(t, ok)

				return nil
			}, sensor)

			lambdacontext.FunctionName = "test-function"
			lambdacontext.FunctionVersion = "42"

			ctx := lambdacontext.NewContext(context.Background(), lc)

			_, err := h.Invoke(ctx, []byte("{}"))
			require.NoError(t, err)

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			span := spans[0]

			assert.EqualValues(t, 0x1234, span.TraceID)
			assert.EqualValues(t, 0x4567, span.ParentID)
			assert.NotEqual(t, span.ParentID, span.SpanID)

			require.Equal(t, "aws.lambda.entry", span.Name)
			assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

			assert.Equal(t, instana.AWSLambdaSpanData{
				Snapshot: instana.AWSLambdaSpanTags{
					ARN:       "aws:test-function:42",
					Runtime:   "go",
					Name:      "test-function",
					Version:   "42",
					Trigger:   "aws:lambda.invoke",
					ColdStart: true,
				},
			}, span.Data)
		})
	}
}

func TestNewHandler_InvokeLambda_ColdStartAndNotColdStart(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instalambda.NewHandler(func(ctx context.Context, evt interface{}) error {
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

	// cold start
	_, err := h.Invoke(ctx, []byte("{}"))
	require.NoError(t, err)

	// second call
	_, err = h.Invoke(ctx, []byte("{}"))
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	assert.Equal(t, instana.AWSLambdaSpanData{
		Snapshot: instana.AWSLambdaSpanTags{
			ARN:       "aws:test-function:42",
			Runtime:   "go",
			Name:      "test-function",
			Version:   "42",
			Trigger:   "aws:lambda.invoke",
			ColdStart: true,
		},
	}, spans[0].Data)

	assert.Equal(t, instana.AWSLambdaSpanData{
		Snapshot: instana.AWSLambdaSpanTags{
			ARN:       "aws:test-function:42",
			Runtime:   "go",
			Name:      "test-function",
			Version:   "42",
			Trigger:   "aws:lambda.invoke",
			ColdStart: false,
		},
	}, spans[1].Data)
}

func TestNewHandler_InvokeLambda_Timeout(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
	defer instana.ShutdownSensor()

	h := instalambda.NewHandler(func() error {
		time.Sleep(100 * time.Millisecond) // make sure the function times out

		return nil
	}, sensor)

	lambdacontext.FunctionName = "test-function"
	lambdacontext.FunctionVersion = "42"

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID:       "req1",
		InvokedFunctionArn: "aws:test-function",
	})

	ctx, cancel := context.WithDeadline(ctx, time.Now())
	defer cancel()

	_, err := h.Invoke(ctx, []byte("{}"))
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	lambdaSpan, logSpan := spans[0], spans[1]

	require.IsType(t, instana.AWSLambdaSpanData{}, lambdaSpan.Data)

	lambdaData := lambdaSpan.Data.(instana.AWSLambdaSpanData)
	assert.Equal(t, instana.AWSLambdaSpanData{
		Snapshot: instana.AWSLambdaSpanTags{
			ARN:              "aws:test-function:42",
			Runtime:          "go",
			Name:             "test-function",
			Version:          "42",
			Trigger:          "aws:lambda.invoke",
			ColdStart:        true,
			MillisecondsLeft: lambdaData.Snapshot.MillisecondsLeft,
		},
	}, lambdaSpan.Data)

	require.IsType(t, instana.LogSpanData{}, logSpan.Data)

	logData := logSpan.Data.(instana.LogSpanData)
	require.Equal(t, "ERROR", logData.Tags.Level)
	require.Equal(t, `error: "handler has timed out"`, logData.Tags.Message)
}

func TestNewHandler_InvokeLambda_WithIncompleteSetOfInstanaHeaders(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(getOptions(), recorder))
	defer instana.ShutdownSensor()

	h := instalambda.NewHandler(func(ctx context.Context, evt interface{}) error {
		_, ok := instana.SpanFromContext(ctx)
		assert.True(t, ok)

		return nil
	}, sensor)

	lambdacontext.FunctionName = "test-function"
	lambdacontext.FunctionVersion = "42"

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID:       "req1",
		InvokedFunctionArn: "aws:test-function",
		ClientContext: lambdacontext.ClientContext{
			Client: lambdacontext.ClientApplication{},
			Env:    nil,
			Custom: map[string]string{
				"WRONG_HEADER_NAME": "0000000000001234",
				instana.FieldS:      "0000000000004567",
				instana.FieldL:      "1",
			},
		},
	})

	_, err := h.Invoke(ctx, []byte("{}"))
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]

	assert.NotEqual(t, 0x1234, span.TraceID)
	assert.NotEqual(t, 0x4567, span.ParentID)
	assert.NotEqual(t, span.ParentID, span.SpanID)

	require.Equal(t, "aws.lambda.entry", span.Name)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)

	assert.Equal(t, instana.AWSLambdaSpanData{
		Snapshot: instana.AWSLambdaSpanTags{
			ARN:       "aws:test-function:42",
			Runtime:   "go",
			Name:      "test-function",
			Version:   "42",
			Trigger:   "aws:lambda.invoke",
			ColdStart: true,
		},
	}, span.Data)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                              { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error  { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error {
	return errors.New("Dummy agent client.")
}
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
