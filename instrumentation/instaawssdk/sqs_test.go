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
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
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
				instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
			)

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
			tracer := instana.NewTracerWithEverything(instana.DefaultOptions(), recorder)

			sp := tracer.StartSpan("sqs")

			req := example.Request()
			req.SetContext(instana.ContextWithSpan(req.Context(), sp))

			instaawssdk.FinalizeSQSSpan(sp, req)

			sp.Finish()

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
	tracer := instana.NewTracerWithEverything(instana.DefaultOptions(), recorder)

	sp := tracer.StartSpan("sqs")

	svc := sqs.New(unit.Session)

	req, _ := svc.GetQueueUrlRequest(&sqs.GetQueueUrlInput{
		QueueName: aws.String("test-queue"),
	})
	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
	req.Error = awserr.New("42", "test error", errors.New("an error occurred"))

	instaawssdk.FinalizeSQSSpan(sp, req)

	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	sqsSpan := spans[0]

	assert.IsType(t, instana.AWSSQSSpanData{}, sqsSpan.Data)
	data := sqsSpan.Data.(instana.AWSSQSSpanData)

	assert.Equal(t, instana.AWSSQSSpanTags{
		Error: "42: test error\ncaused by: an error occurred",
	}, data.Tags)
}
