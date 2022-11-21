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
	"github.com/aws/aws-sdk-go/service/s3"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartS3Span_WithActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	parentSp := sensor.Tracer().StartSpan("testing")

	req := newS3Request()
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartS3Span(req, sensor)

	sp, ok := instana.SpanFromContext(req.Context())
	require.True(t, ok)

	sp.Finish()
	parentSp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	s3Span, testingSpan := spans[0], spans[1]

	assert.Equal(t, testingSpan.TraceID, s3Span.TraceID)
	assert.Equal(t, testingSpan.SpanID, s3Span.ParentID)
	assert.NotEqual(t, testingSpan.SpanID, s3Span.SpanID)
	assert.NotEmpty(t, s3Span.SpanID)

	assert.Equal(t, "s3", s3Span.Name)
	assert.Empty(t, s3Span.Ec)

	assert.IsType(t, instana.AWSS3SpanData{}, s3Span.Data)
	data := s3Span.Data.(instana.AWSS3SpanData)

	assert.Equal(t, instana.AWSS3SpanTags{
		Region:    "mock-region",
		Operation: "put",
		Bucket:    "test-bucket",
		Key:       "test-key",
	}, data.Tags)
}

func TestStartS3Span_NoActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)
	defer instana.ShutdownSensor()

	req := newS3Request()
	instaawssdk.StartS3Span(req, sensor)

	_, ok := instana.SpanFromContext(req.Context())
	require.False(t, ok)
}

func TestFinalizeS3_NoError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("s3", opentracing.Tags{
		"s3.region": "us-east-1",
		"s3.op":     "PutObject",
		"s3.bucket": "test-bucket",
		"s3.key":    "test-key",
	})

	req := newS3Request()
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))

	instaawssdk.FinalizeS3Span(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	s3Span := spans[0]

	assert.IsType(t, instana.AWSS3SpanData{}, s3Span.Data)
	data := s3Span.Data.(instana.AWSS3SpanData)

	assert.Equal(t, instana.AWSS3SpanTags{
		Region:    "us-east-1",
		Operation: "PutObject",
		Bucket:    "test-bucket",
		Key:       "test-key",
	}, data.Tags)
}

func TestFinalizeS3Span_WithError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("s3", opentracing.Tags{
		"s3.region": "us-east-1",
		"s3.op":     "PutObject",
		"s3.bucket": "test-bucket",
		"s3.key":    "test-key",
	})

	req := newS3Request()
	req.Error = awserr.New("42", "test error", errors.New("an error occurred"))
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))

	instaawssdk.FinalizeS3Span(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	s3Span := spans[0]

	assert.IsType(t, instana.AWSS3SpanData{}, s3Span.Data)
	data := s3Span.Data.(instana.AWSS3SpanData)

	assert.Equal(t, instana.AWSS3SpanTags{
		Region:    "us-east-1",
		Operation: "PutObject",
		Bucket:    "test-bucket",
		Key:       "test-key",
		Error:     req.Error.Error(),
	}, data.Tags)
}

func newS3Request() *request.Request {
	svc := s3.New(unit.Session)
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("test-key"),
	})

	return req
}
