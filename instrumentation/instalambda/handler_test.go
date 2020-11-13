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
