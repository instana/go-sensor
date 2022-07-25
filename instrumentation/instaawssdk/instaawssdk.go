// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

// Package instaawssdk instruments github.com/aws/aws-sdk-go
package instaawssdk

import (
	"errors"
	"fmt"

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

// New is a wrapper for `session.New`
func New(sensor *instana.Sensor, cfgs ...*aws.Config) *session.Session {
	sess := session.New(cfgs...)
	InstrumentSession(sess, sensor)

	return sess
}

// NewSession is a wrapper for `session.NewSession`
func NewSession(sensor *instana.Sensor, cfgs ...*aws.Config) (*session.Session, error) {
	sess, err := session.NewSession(cfgs...)
	if err != nil {
		return sess, err
	}

	InstrumentSession(sess, sensor)

	return sess, nil
}

// NewSessionWithOptions is a wrapper for `session.NewSessionWithOptions`
func NewSessionWithOptions(sensor *instana.Sensor, opts session.Options) (*session.Session, error) {
	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return sess, err
	}

	InstrumentSession(sess, sensor)

	return sess, nil
}

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
			StartInvokeLambdaSpan(req, sensor)
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
					sp.SetTag(sqsQueue, aws.StringValue(params.QueueUrl))
					sp.Finish()
				}
			}
		case dynamodb.ServiceName:
			FinalizeDynamoDBSpan(req)
		case lambda.ServiceName:
			FinalizeInvokeLambdaSpan(req)
		}
	})
}

func injectTraceContext(sp opentracing.Span, req *request.Request, logger instana.LeveledLogger) {
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
		var err error
		lc := LambdaClientContext{}

		if params.ClientContext != nil {
			lc, err = NewLambdaClientContextFromBase64EncodedJSON(*params.ClientContext)
			if err != nil {
				logger.Error("lambdaClientContext decode:", err)

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

		s, err := lc.Base64JSON()
		if err != nil {
			logger.Error("lambdaClientContext encode:", err)

			return
		}

		if len(s) <= maxClientContextLen {
			params.ClientContext = &s
		}
	}
}
