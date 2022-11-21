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
	"github.com/aws/aws-sdk-go/service/sns"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartSNSSpan_WithActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	parentSp := sensor.Tracer().StartSpan("testing")

	req := newSNSRequest()
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartSNSSpan(req, sensor)

	sp, ok := instana.SpanFromContext(req.Context())
	require.True(t, ok)

	sp.Finish()
	parentSp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	snsSpan, testingSpan := spans[0], spans[1]

	assert.Equal(t, testingSpan.TraceID, snsSpan.TraceID)
	assert.Equal(t, testingSpan.SpanID, snsSpan.ParentID)
	assert.NotEqual(t, testingSpan.SpanID, snsSpan.SpanID)
	assert.NotEmpty(t, snsSpan.SpanID)

	assert.Equal(t, "sns", snsSpan.Name)
	assert.Empty(t, snsSpan.Ec)

	assert.IsType(t, instana.AWSSNSSpanData{}, snsSpan.Data)
	data := snsSpan.Data.(instana.AWSSNSSpanData)

	assert.Equal(t, instana.AWSSNSSpanTags{
		TopicARN:  "test-topic-arn",
		TargetARN: "test-target-arn",
		Phone:     "test-phone-no",
		Subject:   "test-subject",
	}, data.Tags)
}

func TestStartSNSSpan_NonInstrumentedMethod(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	parentSp := sensor.Tracer().StartSpan("testing")

	svc := sns.New(unit.Session)
	req, _ := svc.CheckIfPhoneNumberIsOptedOutRequest(&sns.CheckIfPhoneNumberIsOptedOutInput{
		PhoneNumber: aws.String("test-phone-no"),
	})
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartSNSSpan(req, sensor)

	sp, ok := instana.SpanFromContext(req.Context())
	assert.True(t, ok)
	assert.Equal(t, parentSp, sp)

	parentSp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
}

func TestStartSNSSpan_TraceContextPropagation_Single(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	svc := sns.New(unit.Session)

	parentSp := sensor.Tracer().StartSpan("testing")

	req, _ := svc.PublishRequest(&sns.PublishInput{
		Message:     aws.String("message-1"),
		PhoneNumber: aws.String("test-phone-no"),
	})
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartSNSSpan(req, sensor)

	sp, ok := instana.SpanFromContext(req.Context())
	require.True(t, ok)

	sp.Finish()
	parentSp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	snsSpan := spans[0]

	params := req.Params.(*sns.PublishInput)
	assert.Equal(t, map[string]*sns.MessageAttributeValue{
		instaawssdk.FieldT: {
			DataType:    aws.String("String"),
			StringValue: aws.String(instana.FormatID(snsSpan.TraceID)),
		},
		instaawssdk.FieldS: {
			DataType:    aws.String("String"),
			StringValue: aws.String(instana.FormatID(snsSpan.SpanID)),
		},
		instaawssdk.FieldL: {
			DataType:    aws.String("String"),
			StringValue: aws.String("1"),
		},
	}, params.MessageAttributes)
}

func TestStartSNSSpan_NoActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)
	defer instana.ShutdownSensor()

	req := newSNSRequest()
	instaawssdk.StartSNSSpan(req, sensor)

	_, ok := instana.SpanFromContext(req.Context())
	require.False(t, ok)
}

func TestFinalizeSNS_NoError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("sns", opentracing.Tags{
		"sns.topic":   "test-topic-arn",
		"sns.target":  "test-target-arn",
		"sns.phone":   "test-phone-no",
		"sns.subject": "test-subject",
	})

	req := newSNSRequest()
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))

	instaawssdk.FinalizeSNSSpan(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	snsSpan := spans[0]

	assert.IsType(t, instana.AWSSNSSpanData{}, snsSpan.Data)
	data := snsSpan.Data.(instana.AWSSNSSpanData)

	assert.Equal(t, instana.AWSSNSSpanTags{
		TopicARN:  "test-topic-arn",
		TargetARN: "test-target-arn",
		Phone:     "test-phone-no",
		Subject:   "test-subject",
	}, data.Tags)
}

func TestFinalizeSNSSpan_WithError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("sns", opentracing.Tags{
		"sns.topic":   "test-topic-arn",
		"sns.target":  "test-target-arn",
		"sns.phone":   "test-phone-no",
		"sns.subject": "test-subject",
	})

	req := newSNSRequest()
	req.Error = awserr.New("42", "test error", errors.New("an error occurred"))
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))

	instaawssdk.FinalizeSNSSpan(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	snsSpan := spans[0]

	assert.IsType(t, instana.AWSSNSSpanData{}, snsSpan.Data)
	data := snsSpan.Data.(instana.AWSSNSSpanData)

	assert.Equal(t, instana.AWSSNSSpanTags{
		TopicARN:  "test-topic-arn",
		TargetARN: "test-target-arn",
		Phone:     "test-phone-no",
		Subject:   "test-subject",
		Error:     req.Error.Error(),
	}, data.Tags)
}

func newSNSRequest() *request.Request {
	svc := sns.New(unit.Session)
	req, _ := svc.PublishRequest(&sns.PublishInput{
		Message:     aws.String("message content"),
		PhoneNumber: aws.String("test-phone-no"),
		Subject:     aws.String("test-subject"),
		TargetArn:   aws.String("test-target-arn"),
		TopicArn:    aws.String("test-topic-arn"),
	})

	return req
}
