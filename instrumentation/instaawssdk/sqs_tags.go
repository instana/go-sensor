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
			sqsQueue: aws.StringValue(params.QueueUrl),
		}, nil
	case *sqs.SendMessageInput:
		return opentracing.Tags{
			sqsType:  "single.sync",
			sqsQueue: aws.StringValue(params.QueueUrl),
			sqsGroup: aws.StringValue(params.MessageGroupId),
		}, nil
	case *sqs.SendMessageBatchInput:
		return opentracing.Tags{
			sqsType:  "batch.sync",
			sqsQueue: aws.StringValue(params.QueueUrl),
			sqsSize:  len(params.Entries),
		}, nil
	case *sqs.GetQueueUrlInput:
		return opentracing.Tags{
			sqsType: "get.queue",
			// the queue url will be returned as a part of response,
			// so we'd need to update this tag once queue is created.
			// however, we keep the name for now in case there will
			// be an error to display the desired name in ui
			sqsQueue: aws.StringValue(params.QueueName),
		}, nil
	case *sqs.CreateQueueInput:
		return opentracing.Tags{
			sqsType: "create.queue",
			// the queue url will be returned as a part of response,
			// so we'd need to update this tag once queue is created.
			// however, we keep the name for now in case there will
			// be an error to display the desired name in ui
			sqsQueue: aws.StringValue(params.QueueName),
		}, nil
	case *sqs.DeleteMessageInput:
		return opentracing.Tags{
			sqsType:  "delete.single.sync",
			sqsQueue: aws.StringValue(params.QueueUrl),
		}, nil
	case *sqs.DeleteMessageBatchInput:
		return opentracing.Tags{
			sqsType:  "delete.batch.sync",
			sqsQueue: aws.StringValue(params.QueueUrl),
			sqsSize:  len(params.Entries),
		}, nil
	default:
		return nil, errMethodNotInstrumented
	}
}
