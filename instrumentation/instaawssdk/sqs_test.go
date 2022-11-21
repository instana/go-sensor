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
	"github.com/aws/aws-sdk-go/service/sqs"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartSQSSpan(t *testing.T) {
	svc := sqs.New(unit.Session)

	examples := map[string]struct {
		Request      func() *request.Request
		ExpectedTags instana.AWSSQSSpanTags
	}{
		"ReceiveMessage": {
			Request: func() *request.Request {
				req, _ := svc.ReceiveMessageRequest(&sqs.ReceiveMessageInput{
					QueueUrl: aws.String("test-queue"),
				})

				return req
			},
			ExpectedTags: instana.AWSSQSSpanTags{
				Sort:  "exit",
				Queue: "test-queue",
			},
		},
		"SendMessage": {
			Request: func() *request.Request {
				req, _ := svc.SendMessageRequest(&sqs.SendMessageInput{
					MessageBody:    aws.String("message-1"),
					MessageGroupId: aws.String("test-group-id"),
					QueueUrl:       aws.String("test-queue"),
				})

				return req
			},
			ExpectedTags: instana.AWSSQSSpanTags{
				Sort:           "exit",
				Type:           "single.sync",
				Queue:          "test-queue",
				MessageGroupID: "test-group-id",
			},
		},
		"SendMessageBatch": {
			Request: func() *request.Request {
				req, _ := svc.SendMessageBatchRequest(&sqs.SendMessageBatchInput{
					Entries: []*sqs.SendMessageBatchRequestEntry{
						{MessageBody: aws.String("message-1")},
						{MessageBody: aws.String("message-2")},
					},
					QueueUrl: aws.String("test-queue"),
				})

				return req
			},
			ExpectedTags: instana.AWSSQSSpanTags{
				Sort:  "exit",
				Type:  "batch.sync",
				Queue: "test-queue",
				Size:  2,
			},
		},
		"GetQueueUrl": {
			Request: func() *request.Request {
				req, _ := svc.GetQueueUrlRequest(&sqs.GetQueueUrlInput{
					QueueName: aws.String("test-queue"),
				})

				return req
			},
			ExpectedTags: instana.AWSSQSSpanTags{
				Sort:  "exit",
				Type:  "get.queue",
				Queue: "test-queue",
			},
		},
		"CreateQueue": {
			Request: func() *request.Request {
				req, _ := svc.CreateQueueRequest(&sqs.CreateQueueInput{
					QueueName: aws.String("test-queue"),
				})

				return req
			},
			ExpectedTags: instana.AWSSQSSpanTags{
				Sort:  "exit",
				Type:  "create.queue",
				Queue: "test-queue",
			},
		},
		"DeleteMessage": {
			Request: func() *request.Request {
				req, _ := svc.DeleteMessageRequest(&sqs.DeleteMessageInput{
					QueueUrl: aws.String("test-queue"),
				})

				return req
			},
			ExpectedTags: instana.AWSSQSSpanTags{
				Sort:  "exit",
				Type:  "delete.single.sync",
				Queue: "test-queue",
			},
		},
		"DeleteMessageBatch": {
			Request: func() *request.Request {
				req, _ := svc.DeleteMessageBatchRequest(&sqs.DeleteMessageBatchInput{
					Entries: []*sqs.DeleteMessageBatchRequestEntry{
						{Id: aws.String("message-1")},
						{Id: aws.String("message-2")},
					},
					QueueUrl: aws.String("test-queue"),
				})

				return req
			},
			ExpectedTags: instana.AWSSQSSpanTags{
				Sort:  "exit",
				Type:  "delete.batch.sync",
				Queue: "test-queue",
				Size:  2,
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
			)
			defer instana.ShutdownSensor()

			parentSp := sensor.Tracer().StartSpan("testing")

			req := example.Request()
			req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

			instaawssdk.StartSQSSpan(req, sensor)

			sp, ok := instana.SpanFromContext(req.Context())
			require.True(t, ok)

			sp.Finish()
			parentSp.Finish()

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 2)

			sqsSpan, testingSpan := spans[0], spans[1]

			assert.Equal(t, testingSpan.TraceID, sqsSpan.TraceID)
			assert.Equal(t, testingSpan.SpanID, sqsSpan.ParentID)
			assert.NotEqual(t, testingSpan.SpanID, sqsSpan.SpanID)
			assert.NotEmpty(t, sqsSpan.SpanID)

			assert.Equal(t, "sqs", sqsSpan.Name)
			assert.Equal(t, int(instana.ExitSpanKind), sqsSpan.Kind)
			assert.Empty(t, sqsSpan.Ec)

			assert.IsType(t, instana.AWSSQSSpanData{}, sqsSpan.Data)
			data := sqsSpan.Data.(instana.AWSSQSSpanData)

			assert.Equal(t, example.ExpectedTags, data.Tags)
		})
	}
}

func TestStartSQSSpan_NonInstrumentedMethod(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	parentSp := sensor.Tracer().StartSpan("testing")

	svc := sqs.New(unit.Session)
	req, _ := svc.RemovePermissionRequest(&sqs.RemovePermissionInput{
		QueueUrl: aws.String("test-queue"),
	})
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartSQSSpan(req, sensor)

	sp, ok := instana.SpanFromContext(req.Context())
	assert.True(t, ok)
	assert.Equal(t, parentSp, sp)

	parentSp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
}

func TestStartSQSSpan_TraceContextPropagation_Single(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	svc := sqs.New(unit.Session)

	parentSp := sensor.Tracer().StartSpan("testing")

	req, _ := svc.SendMessageRequest(&sqs.SendMessageInput{
		MessageBody:    aws.String("message-1"),
		MessageGroupId: aws.String("test-group-id"),
		QueueUrl:       aws.String("test-queue"),
	})
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartSQSSpan(req, sensor)

	sp, ok := instana.SpanFromContext(req.Context())
	require.True(t, ok)

	sp.Finish()
	parentSp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	sqsSpan := spans[0]

	params := req.Params.(*sqs.SendMessageInput)
	assert.Equal(t, map[string]*sqs.MessageAttributeValue{
		instaawssdk.FieldT: {
			DataType:    aws.String("String"),
			StringValue: aws.String(instana.FormatID(sqsSpan.TraceID)),
		},
		instaawssdk.FieldS: {
			DataType:    aws.String("String"),
			StringValue: aws.String(instana.FormatID(sqsSpan.SpanID)),
		},
		instaawssdk.FieldL: {
			DataType:    aws.String("String"),
			StringValue: aws.String("1"),
		},
	}, params.MessageAttributes)
}

func TestStartSQSSpan_TraceContextPropagation_Batch(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	svc := sqs.New(unit.Session)

	parentSp := sensor.Tracer().StartSpan("testing")

	req, _ := svc.SendMessageBatchRequest(&sqs.SendMessageBatchInput{
		Entries: []*sqs.SendMessageBatchRequestEntry{
			{MessageBody: aws.String("message-1"), MessageGroupId: aws.String("test-group-id")},
			{MessageBody: aws.String("message-2"), MessageGroupId: aws.String("test-group-id")},
		},
		QueueUrl: aws.String("test-queue"),
	})
	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))

	instaawssdk.StartSQSSpan(req, sensor)

	sp, ok := instana.SpanFromContext(req.Context())
	require.True(t, ok)

	sp.Finish()
	parentSp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	sqsSpan := spans[0]

	params := req.Params.(*sqs.SendMessageBatchInput)
	for _, entry := range params.Entries {
		assert.Equal(t, map[string]*sqs.MessageAttributeValue{
			instaawssdk.FieldT: {
				DataType:    aws.String("String"),
				StringValue: aws.String(instana.FormatID(sqsSpan.TraceID)),
			},
			instaawssdk.FieldS: {
				DataType:    aws.String("String"),
				StringValue: aws.String(instana.FormatID(sqsSpan.SpanID)),
			},
			instaawssdk.FieldL: {
				DataType:    aws.String("String"),
				StringValue: aws.String("1"),
			},
		}, entry.MessageAttributes)
	}
}

func TestStartSQSSpan_TraceContextPropagation_Single_NoActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)
	defer instana.ShutdownSensor()

	svc := sqs.New(unit.Session)

	req, _ := svc.SendMessageRequest(&sqs.SendMessageInput{
		MessageBody:    aws.String("message-1"),
		MessageGroupId: aws.String("test-group-id"),
		QueueUrl:       aws.String("test-queue"),
	})

	instaawssdk.StartSQSSpan(req, sensor)

	_, ok := instana.SpanFromContext(req.Context())
	require.False(t, ok)

	assert.Empty(t, recorder.GetQueuedSpans())

	params := req.Params.(*sqs.SendMessageInput)

	assert.NotContains(t, params.MessageAttributes, instaawssdk.FieldT)
	assert.NotContains(t, params.MessageAttributes, instaawssdk.FieldS)
	assert.NotContains(t, params.MessageAttributes, instaawssdk.FieldL)
}

func TestStartSQSSpan_TraceContextPropagation_Batch_NoActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)
	defer instana.ShutdownSensor()

	svc := sqs.New(unit.Session)

	req, _ := svc.SendMessageBatchRequest(&sqs.SendMessageBatchInput{
		Entries: []*sqs.SendMessageBatchRequestEntry{
			{MessageBody: aws.String("message-1"), MessageGroupId: aws.String("test-group-id")},
			{MessageBody: aws.String("message-2"), MessageGroupId: aws.String("test-group-id")},
		},
		QueueUrl: aws.String("test-queue"),
	})

	instaawssdk.StartSQSSpan(req, sensor)

	_, ok := instana.SpanFromContext(req.Context())
	require.False(t, ok)

	assert.Empty(t, recorder.GetQueuedSpans())

	params := req.Params.(*sqs.SendMessageBatchInput)
	for _, entry := range params.Entries {
		assert.NotContains(t, entry.MessageAttributes, instaawssdk.FieldT)
		assert.NotContains(t, entry.MessageAttributes, instaawssdk.FieldS)
		assert.NotContains(t, entry.MessageAttributes, instaawssdk.FieldL)
	}
}

func TestFinalizeSQSSpan(t *testing.T) {
	svc := sqs.New(unit.Session)

	examples := map[string]struct {
		Request      func() *request.Request
		ExpectedTags instana.AWSSQSSpanTags
	}{
		"GetQueueUrl": {
			Request: func() *request.Request {
				req, _ := svc.GetQueueUrlRequest(&sqs.GetQueueUrlInput{
					QueueName: aws.String("test-queue"),
				})
				req.Data = &sqs.GetQueueUrlOutput{
					QueueUrl: aws.String("test-queue-url"),
				}

				return req
			},
			ExpectedTags: instana.AWSSQSSpanTags{
				Queue: "test-queue-url",
			},
		},
		"CreateQueue": {
			Request: func() *request.Request {
				req, _ := svc.CreateQueueRequest(&sqs.CreateQueueInput{
					QueueName: aws.String("test-queue"),
				})
				req.Data = &sqs.CreateQueueOutput{
					QueueUrl: aws.String("test-queue-url"),
				}

				return req
			},
			ExpectedTags: instana.AWSSQSSpanTags{
				Queue: "test-queue-url",
			},
		},
		"ReceiveMessage": {
			Request: func() *request.Request {
				req, _ := svc.ReceiveMessageRequest(&sqs.ReceiveMessageInput{
					QueueUrl: aws.String("test-queue"),
				})
				req.Data = &sqs.ReceiveMessageOutput{
					Messages: []*sqs.Message{
						{Body: aws.String("message-1")},
						{Body: aws.String("message-2")},
					},
				}

				return req
			},
			ExpectedTags: instana.AWSSQSSpanTags{
				Size: 2,
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {

			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
			defer instana.ShutdownSensor()

			sp := tracer.StartSpan("sqs")

			req := example.Request()
			req.SetContext(instana.ContextWithSpan(req.Context(), sp))

			instaawssdk.FinalizeSQSSpan(req)

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			sqsSpan := spans[0]

			assert.IsType(t, instana.AWSSQSSpanData{}, sqsSpan.Data)
			data := sqsSpan.Data.(instana.AWSSQSSpanData)

			assert.Equal(t, example.ExpectedTags, data.Tags)
		})
	}
}

func TestFinalizeSQSSpan_WithError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	sp := tracer.StartSpan("sqs")

	svc := sqs.New(unit.Session)

	req, _ := svc.GetQueueUrlRequest(&sqs.GetQueueUrlInput{
		QueueName: aws.String("test-queue"),
	})
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
	req.Error = awserr.New("42", "test error", errors.New("an error occurred"))

	instaawssdk.FinalizeSQSSpan(req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	sqsSpan := spans[0]

	assert.IsType(t, instana.AWSSQSSpanData{}, sqsSpan.Data)
	data := sqsSpan.Data.(instana.AWSSQSSpanData)

	assert.Equal(t, instana.AWSSQSSpanTags{
		Error: "42: test error\ncaused by: an error occurred",
	}, data.Tags)
}

func TestTraceSQSMessage_WithTraceContext(t *testing.T) {
	examples := map[string]*sqs.Message{
		"standard keys": {
			Body: aws.String("message body"),
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"X_INSTANA_T": {
					DataType:    aws.String("String"),
					StringValue: aws.String("00000000000000010000000000000002"),
				},
				"X_INSTANA_S": {
					DataType:    aws.String("String"),
					StringValue: aws.String("0000000000000003"),
				},
				"X_INSTANA_L": {
					DataType:    aws.String("String"),
					StringValue: aws.String("1"),
				},
			},
		},
		"sns notification": {
			Body: aws.String(`{
  "Type" : "Notification",
  "MessageId" : "id1",
  "TopicArn" : "test-topic-arn",
  "Subject" : "Test Message",
  "Message" : "HI MOM!",
  "Timestamp" : "2021-01-27T11:22:17.799Z",
  "SignatureVersion" : "1",
  "MessageAttributes" : {
    "X_INSTANA_T" : {"Type":"String","Value":"00000000000000010000000000000002"},
    "X_INSTANA_S" : {"Type":"String","Value":"0000000000000003"},
    "X_INSTANA_L" : {"Type":"String","Value":"1"}
  }
}`),
		},
	}

	for name, msg := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
			)
			defer instana.ShutdownSensor()

			sp := instaawssdk.TraceSQSMessage(msg, sensor)
			require.Equal(t, 0, recorder.QueuedSpansCount())

			sp.Finish()

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			sqsSpan := spans[0]

			assert.EqualValues(t, 0x1, sqsSpan.TraceIDHi)
			assert.EqualValues(t, 0x2, sqsSpan.TraceID)
			assert.EqualValues(t, 0x3, sqsSpan.ParentID)
			assert.NotEqual(t, sqsSpan.ParentID, sqsSpan.SpanID)

			assert.Equal(t, "sqs", sqsSpan.Name)
			assert.Equal(t, int(instana.EntrySpanKind), sqsSpan.Kind)
			assert.Empty(t, sqsSpan.Ec)

			assert.IsType(t, instana.AWSSQSSpanData{}, sqsSpan.Data)

			data := sqsSpan.Data.(instana.AWSSQSSpanData)
			assert.Equal(t, instana.AWSSQSSpanTags{
				Sort: "entry",
			}, data.Tags)
		})
	}
}

func TestTraceSQSMessage_NoTraceContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	msg := &sqs.Message{
		Body: aws.String("message body"),
	}

	sp := instaawssdk.TraceSQSMessage(msg, sensor)
	require.Equal(t, 0, recorder.QueuedSpansCount())

	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	sqsSpan := spans[0]

	assert.NotEmpty(t, sqsSpan.TraceID)
	assert.Empty(t, sqsSpan.ParentID)
	assert.NotEmpty(t, sqsSpan.SpanID)

	assert.Equal(t, "sqs", sqsSpan.Name)
	assert.Equal(t, int(instana.EntrySpanKind), sqsSpan.Kind)
	assert.Empty(t, sqsSpan.Ec)

	assert.IsType(t, instana.AWSSQSSpanData{}, sqsSpan.Data)

	data := sqsSpan.Data.(instana.AWSSQSSpanData)
	assert.Equal(t, instana.AWSSQSSpanTags{
		Sort: "entry",
	}, data.Tags)
}
