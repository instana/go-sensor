// (c) Copyright IBM Corp. 2023

// Package instaawsv2 provides Instana instrumentation for the aws sdk v2 library.
package instaawsv2

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	smithymiddleware "github.com/aws/smithy-go/middleware"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
)

const (
	// fieldT is the trace ID message attribute key
	fieldT = "X_INSTANA_T"
	// fieldS is the span ID message attribute key
	fieldS = "X_INSTANA_S"
	// fieldL is the trace level message attribute key
	fieldL = "X_INSTANA_L"
)

type AWSOperations interface {
	injectContextWithSpan(instana.TracerLogger, context.Context, interface{}) context.Context
	finishSpan(instana.TracerLogger, context.Context, error)
	injectSpanToCarrier(interface{}, ot.Span) error
	extractTags(interface{}) (ot.Tags, error)
}

func operationById(clientType string) AWSOperations {
	switch clientType {
	case s3.ServiceID:
		return AWSS3Operations{}
	case dynamodb.ServiceID:
		return AWSDynamoDBOperations{}
	case sqs.ServiceID:
		return AWSSQSOperations{}
	case lambda.ServiceID:
		return AWSInvokeLambdaOperations{}
	case sns.ServiceID:
		return AWSSNSOperations{}
	}

	return nil
}

// Instrument adds instana instrumentation to the aws config object
func Instrument(tr instana.TracerLogger, cfg *aws.Config) {
	spanBeginFunc := smithymiddleware.SerializeMiddlewareFunc("InstanaSpanBeginMiddleware", func(
		ctx context.Context, in smithymiddleware.SerializeInput, next smithymiddleware.SerializeHandler) (
		out smithymiddleware.SerializeOutput, metadata smithymiddleware.Metadata, err error) {

		clientType := awsmiddleware.GetServiceID(ctx)

		op := operationById(clientType)

		ctx = op.injectContextWithSpan(tr, ctx, in.Parameters)

		out, metadata, err = next.HandleSerialize(ctx, in)

		return out, metadata, err
	})

	spanEndFunc := smithymiddleware.DeserializeMiddlewareFunc("InstanaSpanEndMiddleware", func(
		ctx context.Context, in smithymiddleware.DeserializeInput, next smithymiddleware.DeserializeHandler) (
		out smithymiddleware.DeserializeOutput, metadata smithymiddleware.Metadata, err error) {

		clientType := awsmiddleware.GetServiceID(ctx)

		op := operationById(clientType)

		out, metadata, err = next.HandleDeserialize(ctx, in)

		tr.Logger().Debug("Identified " + clientType + " operation. Finishing the active span")
		op.finishSpan(tr, ctx, err)

		if clientType == sqs.ServiceID {
			if input, ok := in.Request.(*sqs.ReceiveMessageInput); ok {
				sqsQueueUrl := *input.QueueUrl

				if data, ok := out.Result.(*sqs.ReceiveMessageOutput); ok {

					if _, ok := in.Request.(*sqs.ReceiveMessageInput); !ok {
						tr.Logger().Error("unexpected SQS ReceiveMessage parameters type")
					}

					for i := range data.Messages {
						sp := traceSQSMessage(&data.Messages[i], tr)
						sp.SetTag(sqsQueue, sqsQueueUrl)
						sp.Finish()
					}
				}
			}
		}

		return out, metadata, err
	})

	cfg.APIOptions = append(cfg.APIOptions,
		func(stack *smithymiddleware.Stack) error {
			return stack.Serialize.Add(spanBeginFunc, smithymiddleware.Before)
		},
		func(stack *smithymiddleware.Stack) error {
			return stack.Deserialize.Add(spanEndFunc, smithymiddleware.After)
		})
}

func stringDeRef(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

func stringRef(v string) *string {
	return &v
}

type messageAttributesCarrier struct {
	Attrs interface {
		Get(string) (string, bool)
		Set(string, string)
		Del(string)
	}
}

func (c messageAttributesCarrier) Set(key, val string) {
	switch strings.ToLower(key) {
	case instana.FieldT:
		c.Attrs.Set(fieldT, val)
	case instana.FieldS:
		c.Attrs.Set(fieldS, val)
	case instana.FieldL:
		c.Attrs.Set(fieldL, val)
	}
}

func (c messageAttributesCarrier) ForeachKey(handler func(key, val string) error) error {
	if v, ok := c.Attrs.Get(fieldT); ok {
		handler(instana.FieldT, v)
	}

	if v, ok := c.Attrs.Get(fieldS); ok {
		handler(instana.FieldS, v)
	}

	if v, ok := c.Attrs.Get(fieldL); ok {
		handler(instana.FieldL, v)
	}

	return nil
}
