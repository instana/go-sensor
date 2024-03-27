// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var errUnknownSQSMethod = errors.New("sqs method not instrumented")

type AWSSQSOperations struct{}

var _ AWSOperations = (*AWSSQSOperations)(nil)

func (o AWSSQSOperations) injectContextWithSpan(tr instana.TracerLogger, ctx context.Context, params interface{}) context.Context {
	tags, err := o.extractTags(params)
	if err != nil {
		if errors.Is(err, errUnknownSQSMethod) {
			tr.Logger().Error("failed to identify the sqs method: ", err.Error())
			return ctx
		}
	}

	// An exit span will be created independently without a parent span
	// and sent if the user has opted in.
	opts := []ot.StartSpanOption{
		ext.SpanKindRPCClient,
		ot.Tags{sqsSort: "exit"},
		tags,
	}
	parent, ok := instana.SpanFromContext(ctx)
	if ok {
		opts = append(opts, ot.ChildOf(parent.Context()))
	}
	sp := tr.Tracer().StartSpan("sqs", opts...)

	if err = o.injectSpanToCarrier(params, sp); err != nil {
		tr.Logger().Error("failed to inject span context to the sqs carrier: ", err.Error())
	}

	return instana.ContextWithSpan(ctx, sp)
}

func (o AWSSQSOperations) injectSpanToCarrier(params interface{}, sp ot.Span) error {
	var err error

	switch params := params.(type) {
	case *sqs.SendMessageInput:
		if params.MessageAttributes == nil {
			params.MessageAttributes = make(map[string]types.MessageAttributeValue)
		}

		err = sp.Tracer().Inject(
			sp.Context(),
			ot.TextMap,
			sqsMessageAttributesCarrier(params.MessageAttributes),
		)
	case *sqs.SendMessageBatchInput:
		for i := range params.Entries {
			if params.Entries[i].MessageAttributes == nil {
				params.Entries[i].MessageAttributes = make(map[string]types.MessageAttributeValue)
			}

			err = sp.Tracer().Inject(
				sp.Context(),
				ot.TextMap,
				sqsMessageAttributesCarrier(params.Entries[i].MessageAttributes),
			)
		}
	}
	return err
}

func (o AWSSQSOperations) finishSpan(tr instana.TracerLogger, ctx context.Context, err error) {
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

func (o AWSSQSOperations) extractTags(params interface{}) (ot.Tags, error) {
	switch params := params.(type) {
	case *sqs.ReceiveMessageInput:
		return ot.Tags{
			sqsQueue: stringDeRef(params.QueueUrl),
		}, nil
	case *sqs.SendMessageInput:
		return ot.Tags{
			sqsType:  "single.sync",
			sqsQueue: stringDeRef(params.QueueUrl),
			sqsGroup: stringDeRef(params.MessageGroupId),
		}, nil
	case *sqs.SendMessageBatchInput:
		return ot.Tags{
			sqsType:  "batch.sync",
			sqsQueue: stringDeRef(params.QueueUrl),
			sqsSize:  len(params.Entries),
		}, nil
	case *sqs.GetQueueUrlInput:
		return ot.Tags{
			sqsType: "get.queue",
			// the queue url will be returned as a part of response,
			// so we'd need to update this tag once queue is created.
			// however, we keep the name for now in case there will
			// be an error to display the desired name in ui
			sqsQueue: stringDeRef(params.QueueName),
		}, nil
	case *sqs.CreateQueueInput:
		return ot.Tags{
			sqsType: "create.queue",
			// the queue url will be returned as a part of response,
			// so we'd need to update this tag once queue is created.
			// however, we keep the name for now in case there will
			// be an error to display the desired name in ui
			sqsQueue: stringDeRef(params.QueueName),
		}, nil
	case *sqs.DeleteMessageInput:
		return ot.Tags{
			sqsType:  "delete.single.sync",
			sqsQueue: stringDeRef(params.QueueUrl),
		}, nil
	case *sqs.DeleteMessageBatchInput:
		return ot.Tags{
			sqsType:  "delete.batch.sync",
			sqsQueue: stringDeRef(params.QueueUrl),
			sqsSize:  len(params.Entries),
		}, nil
	default:
		return nil, errUnknownSQSMethod
	}
}
