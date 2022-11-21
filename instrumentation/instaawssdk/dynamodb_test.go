// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartDynamoDBSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	parentSp := sensor.Tracer().StartSpan("testing")

	req := newDynamoDBRequest()
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartDynamoDBSpan(req, sensor)

	sp, ok := instana.SpanFromContext(req.Context())
	require.True(t, ok)

	sp.Finish()
	parentSp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	dbSpan, testingSpan := spans[0], spans[1]

	assert.Equal(t, testingSpan.TraceID, dbSpan.TraceID)
	assert.Equal(t, testingSpan.SpanID, dbSpan.ParentID)
	assert.NotEqual(t, testingSpan.SpanID, dbSpan.SpanID)
	assert.NotEmpty(t, dbSpan.SpanID)

	assert.Equal(t, "dynamodb", dbSpan.Name)
	assert.Empty(t, dbSpan.Ec)

	assert.IsType(t, instana.AWSDynamoDBSpanData{}, dbSpan.Data)
	data := dbSpan.Data.(instana.AWSDynamoDBSpanData)

	assert.Equal(t, instana.AWSDynamoDBSpanTags{
		Operation: "get",
		Table:     "test-table",
		Region:    "mock-region",
	}, data.Tags)
}

func TestStartDynamoDBSpan_NonInstrumentedMethod(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	parentSp := sensor.Tracer().StartSpan("testing")

	svc := dynamodb.New(unit.Session)
	req, _ := svc.DescribeLimitsRequest(&dynamodb.DescribeLimitsInput{})
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartDynamoDBSpan(req, sensor)

	sp, ok := instana.SpanFromContext(req.Context())
	require.True(t, ok)

	assert.Equal(t, parentSp, sp)

	parentSp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
}

func TestStartDynamoDBSpan_NoActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)
	defer instana.ShutdownSensor()

	req := newDynamoDBRequest()
	instaawssdk.StartDynamoDBSpan(req, sensor)

	_, ok := instana.SpanFromContext(req.Context())
	require.False(t, ok)
}

func TestFinalizeDynamoDB_NoError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("dynamodb", opentracing.Tags{
		"dynamodb.op":    "get",
		"dynamodb.table": "test-table",
	})

	req := newDynamoDBRequest()
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))

	instaawssdk.FinalizeDynamoDBSpan(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	dbSpan := spans[0]

	assert.IsType(t, instana.AWSDynamoDBSpanData{}, dbSpan.Data)
	data := dbSpan.Data.(instana.AWSDynamoDBSpanData)

	assert.Equal(t, instana.AWSDynamoDBSpanTags{
		Operation: "get",
		Table:     "test-table",
	}, data.Tags)
}

func TestFinalizeDynamoDBSpan_WithError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("dynamodb", opentracing.Tags{
		"dynamodb.op":     "get",
		"dynamodb.table":  "test-table",
		"dynamodb.region": "mock-region",
	})

	req := newDynamoDBRequest()
	req.Error = awserr.New("42", "test error", errors.New("an error occurred"))
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))

	instaawssdk.FinalizeDynamoDBSpan(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	dbSpan := spans[0]

	assert.IsType(t, instana.AWSDynamoDBSpanData{}, dbSpan.Data)
	data := dbSpan.Data.(instana.AWSDynamoDBSpanData)

	assert.Equal(t, instana.AWSDynamoDBSpanTags{
		Operation: "get",
		Table:     "test-table",
		Error:     req.Error.Error(),
		Region:    "mock-region",
	}, data.Tags)
}

func newDynamoDBRequest() *request.Request {
	svc := dynamodb.New(unit.Session)
	req, _ := svc.GetItemRequest(&dynamodb.GetItemInput{
		TableName: aws.String("test-table"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String("id1")},
		},
	})

	return req
}
