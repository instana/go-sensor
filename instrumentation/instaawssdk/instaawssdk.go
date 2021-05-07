// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

// Package instaawssdk instruments github.com/aws/aws-sdk-go

package instaawssdk

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
)

var errMethodNotInstrumented = errors.New("method not instrumented")

const maxClientContextLen = 3582

// InstrumentSession instruments github.com/aws/aws-sdk-go/aws/session.Session by
// injecting handlers to create and finalize Instana spans
func InstrumentSession(sess *session.Session, sensor *instana.Sensor) {
	sess.Handlers.Validate.PushBack(func(req *request.Request) {
		switch req.ClientInfo.ServiceName {
		case s3.ServiceName:
			StartS3Span(req, sensor)
		case sqs.ServiceName:
			StartSQSSpan(req, sensor)
		case dynamodb.ServiceName:
			StartDynamoDBSpan(req, sensor)
		case lambda.ServiceName:
			StartInvokeSpan(req, sensor)
		}
	})

	sess.Handlers.Complete.PushBack(func(req *request.Request) {
		switch req.ClientInfo.ServiceName {
		case s3.ServiceName:
			FinalizeS3Span(req)
		case sqs.ServiceName:
			FinalizeSQSSpan(req)

			if data, ok := req.Data.(*sqs.ReceiveMessageOutput); ok {
				params, ok := req.Params.(*sqs.ReceiveMessageInput)
				if !ok {
					sensor.Logger().Error(fmt.Sprintf("unexpected SQS ReceiveMessage parameters type: %T", req.Params))
					break
				}

				for i := range data.Messages {
					sp := TraceSQSMessage(data.Messages[i], sensor)
					sp.SetTag("sqs.queue", aws.StringValue(params.QueueUrl))
					sp.Finish()
				}
			}
		case dynamodb.ServiceName:
			FinalizeDynamoDBSpan(req)
		case lambda.ServiceName:
			FinalizeInvokeSpan(req)
		}
	})
}

func injectTraceContext(sp opentracing.Span, req *request.Request) {
	switch params := req.Params.(type) {
	case *sqs.SendMessageInput:
		if params.MessageAttributes == nil {
			params.MessageAttributes = make(map[string]*sqs.MessageAttributeValue)
		}

		sp.Tracer().Inject(
			sp.Context(),
			opentracing.TextMap,
			SQSMessageAttributesCarrier(params.MessageAttributes),
		)
	case *sqs.SendMessageBatchInput:
		for i := range params.Entries {
			if params.Entries[i].MessageAttributes == nil {
				params.Entries[i].MessageAttributes = make(map[string]*sqs.MessageAttributeValue)
			}

			sp.Tracer().Inject(
				sp.Context(),
				opentracing.TextMap,
				SQSMessageAttributesCarrier(params.Entries[i].MessageAttributes),
			)
		}
	case *sns.PublishInput:
		if params.MessageAttributes == nil {
			params.MessageAttributes = make(map[string]*sns.MessageAttributeValue)
		}

		sp.Tracer().Inject(
			sp.Context(),
			opentracing.TextMap,
			SNSMessageAttributesCarrier(params.MessageAttributes),
		)
	case *lambda.InvokeInput:
		lc := LambdaClientContext{}

		if params.ClientContext != nil {
			res, err := base64.StdEncoding.DecodeString(*params.ClientContext)
			if err != nil {
				sp.LogFields(otlog.Error(req.Error))

				return
			}

			err = json.Unmarshal(res, &lc)
			if err != nil {
				sp.LogFields(otlog.Error(req.Error))

				return
			}
		}

		if lc.Custom == nil {
			lc.Custom = make(map[string]string)
		}

		sp.Tracer().Inject(
			sp.Context(),
			opentracing.TextMap,
			opentracing.TextMapCarrier(lc.Custom),
		)

		s, err := encodeToBase64(lc)
		if err != nil {
			sp.LogFields(otlog.Error(req.Error))

			return
		}

		if len(s) <= maxClientContextLen {
			params.ClientContext = &s
		}
	}
}
