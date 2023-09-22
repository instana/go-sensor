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
)

const (
	// fieldT is the trace ID message attribute key
	fieldT = "X_INSTANA_T"
	// fieldS is the span ID message attribute key
	fieldS = "X_INSTANA_S"
	// fieldL is the trace level message attribute key
	fieldL = "X_INSTANA_L"
)

// Instrument adds instana instrumentation to the aws config object
func Instrument(tr instana.TracerLogger, cfg *aws.Config) {
	spanBeginFunc := smithymiddleware.SerializeMiddlewareFunc("InstanaSpanBeginMiddleware", func(
		ctx context.Context, in smithymiddleware.SerializeInput, next smithymiddleware.SerializeHandler) (
		out smithymiddleware.SerializeOutput, metadata smithymiddleware.Metadata, err error) {

		clientType := awsmiddleware.GetServiceID(ctx)

		switch clientType {

		case s3.ServiceID:
			tr.Logger().Debug("Identified s3 operation. Initiating s3 span creation")
			ctx = injectAWSContextWithS3Span(tr, ctx, in.Parameters)

		case dynamodb.ServiceID:
			tr.Logger().Debug("Identified dynamodb operation. Initiating dynamodb span creation")
			ctx = injectAWSContextWithDynamoDBSpan(tr, ctx, in.Parameters)

		case sqs.ServiceID:
			tr.Logger().Debug("Identified sqs operation. Initiating sqs span creation")
			ctx = injectAWSContextWithSQSSpan(tr, ctx, in.Parameters)

		case lambda.ServiceID:
			tr.Logger().Debug("Identified lambda operation. Initiating lambda span creation")
			ctx = injectAWSContextWithInvokeLambdaSpan(tr, ctx, in.Parameters)

		case sns.ServiceID:
			tr.Logger().Debug("Identified sns operation. Initiating sns span creation")
			ctx = injectAWSContextWithSNSSpan(tr, ctx, in.Parameters)
		}

		out, metadata, err = next.HandleSerialize(ctx, in)

		return out, metadata, err
	})

	spanEndFunc := smithymiddleware.DeserializeMiddlewareFunc("InstanaSpanEndMiddleware", func(
		ctx context.Context, in smithymiddleware.DeserializeInput, next smithymiddleware.DeserializeHandler) (
		out smithymiddleware.DeserializeOutput, metadata smithymiddleware.Metadata, err error) {

		clientType := awsmiddleware.GetServiceID(ctx)

		var sqsQueueUrl string

		switch clientType {
		case sqs.ServiceID:
			if input, ok := in.Request.(*sqs.ReceiveMessageInput); ok {
				sqsQueueUrl = *input.QueueUrl
			}
		}

		out, metadata, err = next.HandleDeserialize(ctx, in)

		switch clientType {

		case s3.ServiceID:
			tr.Logger().Debug("Identified s3 operation. Finishing the active s3 span")
			finishS3Span(tr, ctx, err)

		case dynamodb.ServiceID:
			tr.Logger().Debug("Identified dynamodb operation. Finishing the active dynamodb span")
			finishDynamoDBSpan(tr, ctx, err)

		case sqs.ServiceID:
			tr.Logger().Debug("Identified sqs operation. Finishing the active sqs span")
			finishSQSSpan(tr, ctx, err)

			if data, ok := out.Result.(*sqs.ReceiveMessageOutput); ok {

				if _, ok := in.Request.(*sqs.ReceiveMessageInput); !ok {
					tr.Logger().Error("unexpected SQS ReceiveMessage parameters type")
					break
				}

				for i := range data.Messages {
					sp := traceSQSMessage(&data.Messages[i], tr)
					sp.SetTag(sqsQueue, sqsQueueUrl)
					sp.Finish()
				}
			}

		case lambda.ServiceID:
			tr.Logger().Debug("Identified lambda operation. Finishing the active lambda span")
			finishInvokeLambdaSpan(tr, ctx, err)

		case sns.ServiceID:
			tr.Logger().Debug("Identified sns operation. Finishing the active sns span")
			finishSNSSpan(tr, ctx, err)
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
	var err error
	if v, ok := c.Attrs.Get(fieldT); ok {
		err = handler(instana.FieldT, v)
	}

	if v, ok := c.Attrs.Get(fieldS); ok {
		err = handler(instana.FieldS, v)
	}

	if v, ok := c.Attrs.Get(fieldL); ok {
		err = handler(instana.FieldL, v)
	}

	return err
}
