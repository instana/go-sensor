// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// StartSQSSpan initiates a new span from an AWS SQS request and injects it into the
// request.Request context
func StartSQSSpan(req *request.Request, sensor *instana.Sensor) {
	tags, err := extractSQSTags(req)
	if err != nil {
		if err == errMethodNotInstrumented {
			return
		}

		sensor.Logger().Warn("failed to extract SQS tags: ", err)
	}

	parent, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}

	sp := sensor.Tracer().StartSpan("sqs",
		ext.SpanKindProducer,
		opentracing.ChildOf(parent.Context()),
		opentracing.Tags{sqsSort: "exit"},
		tags,
	)

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
	injectTraceContext(sp, req, sensor.Logger())
}

// FinalizeSQSSpan retrieves tags from completed request.Request and adds them
// to the span
func FinalizeSQSSpan(req *request.Request) {
	sp, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}
	defer sp.Finish()

	if req.Error != nil {
		sp.LogFields(otlog.Error(req.Error))
		sp.SetTag(sqsError, req.Error.Error())
	}

	if req.DataFilled() {
		switch data := req.Data.(type) {
		case *sqs.GetQueueUrlOutput:
			sp.SetTag(sqsQueue, aws.StringValue(data.QueueUrl))
		case *sqs.CreateQueueOutput:
			sp.SetTag(sqsQueue, aws.StringValue(data.QueueUrl))
		case *sqs.ReceiveMessageOutput:
			sp.SetTag(sqsSize, len(data.Messages))
		}
	}
}

// TraceSQSMessage creates an returns an entry span for an SQS message. The context of this span is injected
// into message attributes. This context can than be retrieved with instaawssdk.SpanContextFromSQSMessage()
// and used in the message handler method to continue the trace.
func TraceSQSMessage(msg *sqs.Message, sensor *instana.Sensor) opentracing.Span {
	opts := []opentracing.StartSpanOption{
		ext.SpanKindConsumer,
		opentracing.Tags{
			sqsSort: "entry",
		},
	}

	if spCtx, ok := SpanContextFromSQSMessage(msg, sensor); ok {
		opts = append(opts, opentracing.ChildOf(spCtx))
	} else {
		body := aws.StringValue(msg.Body)

		// In case the delivery has been created via a subscription to an SNS topic,
		// the message body will be a JSON document containing the SNS notification
		// along with message attributes.
		var payload struct {
			MessageAttributes snsMessageAttributes `json:"MessageAttributes"`
		}

		// try to unmarshal the message attributes and extract the trace context from there
		if err := json.Unmarshal([]byte(body), &payload); err == nil {
			if spCtx, err := sensor.Tracer().Extract(
				opentracing.TextMap,
				SNSMessageAttributesCarrier(map[string]*sns.MessageAttributeValue(payload.MessageAttributes)),
			); err == nil {
				opts = append(opts, opentracing.ChildOf(spCtx))
			}
		}
	}

	sp := sensor.Tracer().StartSpan("sqs", opts...)

	if msg.MessageAttributes == nil {
		msg.MessageAttributes = make(map[string]*sqs.MessageAttributeValue)
	}

	sp.Tracer().Inject(
		sp.Context(),
		opentracing.TextMap,
		SQSMessageAttributesCarrier(msg.MessageAttributes),
	)

	return sp
}
