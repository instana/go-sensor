// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/opentracing/opentracing-go"
)

func extractSQSTags(req *request.Request) opentracing.Tags {
	switch params := req.Params.(type) {
	case *sqs.ReceiveMessageInput:
		return opentracing.Tags{
			"sqs.queue": aws.StringValue(params.QueueUrl),
		}
	case *sqs.SendMessageInput:
		return opentracing.Tags{
			"sqs.queue": aws.StringValue(params.QueueUrl),
			"sqs.group": aws.StringValue(params.MessageGroupId),
		}
	case *sqs.SendMessageBatchInput:
		return opentracing.Tags{
			"sqs.queue": aws.StringValue(params.QueueUrl),
			"sqs.size":  len(params.Entries),
		}
	case *sqs.GetQueueUrlInput:
		return opentracing.Tags{
			// the queue url will be returned as a part of response,
			// so we'd need to update this tag once queue is created.
			// however, we keep the name for now in case there will
			// be an error to display the desired name in ui
			"sqs.queue": aws.StringValue(params.QueueName),
		}
	case *sqs.CreateQueueInput:
		return opentracing.Tags{
			// the queue url will be returned as a part of response,
			// so we'd need to update this tag once queue is created.
			// however, we keep the name for now in case there will
			// be an error to display the desired name in ui
			"sqs.queue": aws.StringValue(params.QueueName),
		}
	case *sqs.DeleteMessageInput:
		return opentracing.Tags{
			"sqs.queue": aws.StringValue(params.QueueUrl),
		}
	case *sqs.DeleteMessageBatchInput:
		return opentracing.Tags{
			"sqs.queue": aws.StringValue(params.QueueUrl),
			"sqs.size":  len(params.Entries),
		}
	default:
		return nil
	}
}
