// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawsv2"
	"github.com/stretchr/testify/assert"
)

func TestSendMessageSQS(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	ctx := context.Background()
	ps := c.Tracer().StartSpan("aws-parent-span")
	ctx = instana.ContextWithSpan(ctx, ps)

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err, "Error while configuring aws")

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(c, &cfg)

	sqsClient := sqs.NewFromConfig(cfg)

	sqsUrl := "test-url"
	sqsMsg := "this is is a test message"
	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody:  &sqsMsg,
		QueueUrl:     &sqsUrl,
		DelaySeconds: 0,
	})
	assert.NoError(t, err, "Error while publishing the sqs message")

	ps.Finish()

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 2, len(recorderSpans))

	sqsSpan := recorderSpans[0]
	assert.IsType(t, instana.AWSSQSSpanData{}, sqsSpan.Data)

	data := sqsSpan.Data.(instana.AWSSQSSpanData)
	assert.Equal(t, instana.AWSSQSSpanTags{
		Queue: sqsUrl,
		Sort:  "exit",
		Type:  "single.sync",
		Error: "",
	}, data.Tags)
}

func TestSQSMonitoredFunctions(t *testing.T) {
	sqsUrl := "test-sqs-url"
	sqsReceiptHandle := "test-receipt-handle"
	sqsBatchRequestEntries := make([]types.SendMessageBatchRequestEntry, 0)
	sqsBatchRequestEntries = append(sqsBatchRequestEntries, types.SendMessageBatchRequestEntry{
		Id:          testString(10),
		MessageBody: testString(30),
	})
	sqsBatchRequestEntries = append(sqsBatchRequestEntries, types.SendMessageBatchRequestEntry{
		Id:          testString(10),
		MessageBody: testString(30),
	})

	testcases := map[string]struct {
		monitoredFunc func(ctx context.Context, client *sqs.Client) (interface{}, error)
		expectedOut   instana.AWSSQSSpanTags
	}{
		"ReceiveMessage": {
			monitoredFunc: func(ctx context.Context, client *sqs.Client) (interface{}, error) {
				ip := sqs.ReceiveMessageInput{
					QueueUrl: &sqsUrl,
				}

				return client.ReceiveMessage(ctx, &ip)
			},
			expectedOut: instana.AWSSQSSpanTags{
				Queue: sqsUrl,
				Sort:  "exit",
				Error: "",
			},
		},
		"SendMessageBatch": {
			monitoredFunc: func(ctx context.Context, client *sqs.Client) (interface{}, error) {
				ip := sqs.SendMessageBatchInput{
					QueueUrl: &sqsUrl,
					Entries:  sqsBatchRequestEntries,
				}

				return client.SendMessageBatch(ctx, &ip)
			},
			expectedOut: instana.AWSSQSSpanTags{
				Queue: sqsUrl,
				Sort:  "exit",
				Type:  "batch.sync",
				Size:  2,
			},
		},
		"GetQueueUrl": {
			monitoredFunc: func(ctx context.Context, client *sqs.Client) (interface{}, error) {
				ip := sqs.GetQueueUrlInput{
					QueueName: &sqsUrl,
				}

				return client.GetQueueUrl(ctx, &ip)
			},
			expectedOut: instana.AWSSQSSpanTags{
				Queue: sqsUrl,
				Sort:  "exit",
				Type:  "get.queue",
			},
		},
		"CreateQueue": {
			monitoredFunc: func(ctx context.Context, client *sqs.Client) (interface{}, error) {
				ip := sqs.CreateQueueInput{
					QueueName: &sqsUrl,
				}

				return client.CreateQueue(ctx, &ip)
			},
			expectedOut: instana.AWSSQSSpanTags{
				Queue: sqsUrl,
				Sort:  "exit",
				Type:  "create.queue",
			},
		},
		"DeleteMessage": {
			monitoredFunc: func(ctx context.Context, client *sqs.Client) (interface{}, error) {
				ip := sqs.DeleteMessageInput{
					QueueUrl:      &sqsUrl,
					ReceiptHandle: &sqsReceiptHandle,
				}

				return client.DeleteMessage(ctx, &ip)
			},
			expectedOut: instana.AWSSQSSpanTags{
				Queue: sqsUrl,
				Sort:  "exit",
				Type:  "delete.single.sync",
			},
		},
		"DeleteMessageBatch": {
			monitoredFunc: func(ctx context.Context, client *sqs.Client) (interface{}, error) {
				ip := sqs.DeleteMessageBatchInput{
					QueueUrl: &sqsUrl,
					Entries:  make([]types.DeleteMessageBatchRequestEntry, 0),
				}

				return client.DeleteMessageBatch(ctx, &ip)
			},
			expectedOut: instana.AWSSQSSpanTags{
				Queue: sqsUrl,
				Sort:  "exit",
				Type:  "delete.batch.sync",
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			c := instana.InitCollector(&instana.Options{
				AgentClient: alwaysReadyClient{},
				Recorder:    recorder,
			})
			defer instana.ShutdownCollector()

			ctx := context.Background()
			ps := c.Tracer().StartSpan("aws-parent-span")
			ctx = instana.ContextWithSpan(ctx, ps)

			cfg, err := config.LoadDefaultConfig(ctx)
			assert.NoError(t, err, "Error while configuring aws")

			cfg = applyTestingChanges(cfg)

			instaawsv2.Instrument(c, &cfg)

			sqsClient := sqs.NewFromConfig(cfg)

			_, err = testcase.monitoredFunc(ctx, sqsClient)

			assert.NoError(t, err, "Error while publishing the sqs message")

			ps.Finish()

			recorderSpans := recorder.GetQueuedSpans()
			assert.Equal(t, 2, len(recorderSpans))

			sqsSpan := recorderSpans[0]
			assert.IsType(t, instana.AWSSQSSpanData{}, sqsSpan.Data)

			data := sqsSpan.Data.(instana.AWSSQSSpanData)
			assert.Equal(t, testcase.expectedOut, data.Tags)
		})
	}

}

func TestSendNoParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err, "Error while configuring aws")

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(c, &cfg)

	sqsClient := sqs.NewFromConfig(cfg)

	sqsUrl := "test-url"
	_, err = sqsClient.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: &sqsUrl,
		Entries:  make([]types.SendMessageBatchRequestEntry, 0),
	})

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 0, len(recorderSpans))
}

func TestSQSUnMonitoredFunction(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err, "Error while configuring aws")

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(c, &cfg)

	sqsClient := sqs.NewFromConfig(cfg)

	queueUrl := "test-queue-url"
	_, err = sqsClient.DeleteQueue(ctx, &sqs.DeleteQueueInput{
		QueueUrl: &queueUrl,
	})

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 0, len(recorderSpans))
}
