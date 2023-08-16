// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var errUnknownSQSMethod = errors.New("sqs method not instrumented")

func injectAWSContextWithSQSSpan(tr instana.TracerLogger, ctx context.Context, params interface{}) context.Context {
	tags, err := extractSQSTags(params)
	if err != nil {
		if errors.Is(err, errUnknownSQSMethod) {
			tr.Logger().Error("failed to identify the sqs method: ", err.Error())
			return ctx
		}
	}

	// By design, we will abort the sqs span creation if a parent span is not identified.
	parent, ok := instana.SpanFromContext(ctx)
	if !ok {
		tr.Logger().Error("failed to retrieve the parent span. Aborting sqs child span creation.")
		return ctx
	}

	sp := tr.Tracer().StartSpan("sqs",
		ext.SpanKindRPCClient,
		opentracing.ChildOf(parent.Context()),
		opentracing.Tags{sqsSort: "exit"},
		tags,
	)

	injectSpanToCarrier(params, sp)

	return instana.ContextWithSpan(ctx, sp)
}

func injectSpanToCarrier(params interface{}, sp opentracing.Span) {
	switch params := params.(type) {
	case *sqs.SendMessageInput:
		if params.MessageAttributes == nil {
			params.MessageAttributes = make(map[string]types.MessageAttributeValue)
		}

		sp.Tracer().Inject(
			sp.Context(),
			opentracing.TextMap,
			sqsMessageAttributesCarrier(params.MessageAttributes),
		)
	case *sqs.SendMessageBatchInput:
		for i := range params.Entries {
			if params.Entries[i].MessageAttributes == nil {
				params.Entries[i].MessageAttributes = make(map[string]types.MessageAttributeValue)
			}

			sp.Tracer().Inject(
				sp.Context(),
				opentracing.TextMap,
				sqsMessageAttributesCarrier(params.Entries[i].MessageAttributes),
			)
		}
	}

}

func finishSQSSpan(tr instana.TracerLogger, ctx context.Context, err error) {
	sp, ok := instana.SpanFromContext(ctx)
	if !ok {
		tr.Logger().Error("failed to retrieve the sqs child span from context.")
		return
	}
	defer sp.Finish()

	if err != nil {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(sqsError, err.Error())
	}
}

func extractSQSTags(params interface{}) (opentracing.Tags, error) {
	switch params := params.(type) {
	case *sqs.ReceiveMessageInput:
		return opentracing.Tags{
			sqsQueue: stringDeRef(params.QueueUrl),
		}, nil
	case *sqs.SendMessageInput:
		return opentracing.Tags{
			sqsType:  "single.sync",
			sqsQueue: stringDeRef(params.QueueUrl),
			sqsGroup: stringDeRef(params.MessageGroupId),
		}, nil
	case *sqs.SendMessageBatchInput:
		return opentracing.Tags{
			sqsType:  "batch.sync",
			sqsQueue: stringDeRef(params.QueueUrl),
			sqsSize:  len(params.Entries),
		}, nil
	case *sqs.GetQueueUrlInput:
		return opentracing.Tags{
			sqsType: "get.queue",
			// the queue url will be returned as a part of response,
			// so we'd need to update this tag once queue is created.
			// however, we keep the name for now in case there will
			// be an error to display the desired name in ui
			sqsQueue: stringDeRef(params.QueueName),
		}, nil
	case *sqs.CreateQueueInput:
		return opentracing.Tags{
			sqsType: "create.queue",
			// the queue url will be returned as a part of response,
			// so we'd need to update this tag once queue is created.
			// however, we keep the name for now in case there will
			// be an error to display the desired name in ui
			sqsQueue: stringDeRef(params.QueueName),
		}, nil
	case *sqs.DeleteMessageInput:
		return opentracing.Tags{
			sqsType:  "delete.single.sync",
			sqsQueue: stringDeRef(params.QueueUrl),
		}, nil
	case *sqs.DeleteMessageBatchInput:
		return opentracing.Tags{
			sqsType:  "delete.batch.sync",
			sqsQueue: stringDeRef(params.QueueUrl),
			sqsSize:  len(params.Entries),
		}, nil
	default:
		return nil, errUnknownSQSMethod
	}
}
