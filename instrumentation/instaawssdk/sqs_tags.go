// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/opentracing/opentracing-go"
)

func extractSQSTags(req *request.Request) (opentracing.Tags, error) {
	switch params := req.Params.(type) {
	case *sqs.ReceiveMessageInput:
		return opentracing.Tags{
			"sqs.queue": aws.StringValue(params.QueueUrl),
		}, nil
	case *sqs.SendMessageInput:
		return opentracing.Tags{
			"sqs.type":  "single.sync",
			"sqs.queue": aws.StringValue(params.QueueUrl),
			"sqs.group": aws.StringValue(params.MessageGroupId),
		}, nil
	case *sqs.SendMessageBatchInput:
		return opentracing.Tags{
			"sqs.type":  "batch.sync",
			"sqs.queue": aws.StringValue(params.QueueUrl),
			"sqs.size":  len(params.Entries),
		}, nil
	case *sqs.GetQueueUrlInput:
		return opentracing.Tags{
			"sqs.type": "get.queue",
			// the queue url will be returned as a part of response,
			// so we'd need to update this tag once queue is created.
			// however, we keep the name for now in case there will
			// be an error to display the desired name in ui
			"sqs.queue": aws.StringValue(params.QueueName),
		}, nil
	case *sqs.CreateQueueInput:
		return opentracing.Tags{
			"sqs.type": "create.queue",
			// the queue url will be returned as a part of response,
			// so we'd need to update this tag once queue is created.
			// however, we keep the name for now in case there will
			// be an error to display the desired name in ui
			"sqs.queue": aws.StringValue(params.QueueName),
		}, nil
	case *sqs.DeleteMessageInput:
		return opentracing.Tags{
			"sqs.type":  "delete.single.sync",
			"sqs.queue": aws.StringValue(params.QueueUrl),
		}, nil
	case *sqs.DeleteMessageBatchInput:
		return opentracing.Tags{
			"sqs.type":  "delete.batch.sync",
			"sqs.queue": aws.StringValue(params.QueueUrl),
			"sqs.size":  len(params.Entries),
		}, nil
	default:
		return nil, errMethodNotInstrumented
	}
}
