// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var sqsInstrumentedOps = map[string]string{
	"ReceiveMessage":     "",
	"SendMessage":        "single.sync",
	"SendMessageBatch":   "batch.sync",
	"GetQueueUrl":        "get.queue",
	"CreateQueue":        "create.queue",
	"DeleteMessage":      "delete.single.sync",
	"DeleteMessageBatch": "delete.batch.sync",
}

// StartSQSSpan initiates a new span from an AWS SQS request and injects it into the
// request.Request context
func StartSQSSpan(req *request.Request, sensor *instana.Sensor) {
	op, ok := sqsInstrumentedOps[req.Operation.Name]
	if !ok {
		return
	}

	startSQSExitSpan(op, req, sensor)
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
		sp.SetTag("sqs.error", req.Error.Error())
	}

	if req.DataFilled() {
		switch data := req.Data.(type) {
		case *sqs.GetQueueUrlOutput:
			sp.SetTag("sqs.queue", aws.StringValue(data.QueueUrl))
		case *sqs.CreateQueueOutput:
			sp.SetTag("sqs.queue", aws.StringValue(data.QueueUrl))
		case *sqs.ReceiveMessageOutput:
			sp.SetTag("sqs.size", len(data.Messages))
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
			"sqs.sort": "entry",
		},
	}

	if spCtx, ok := SpanContextFromSQSMessage(msg, sensor); ok {
		opts = append(opts, opentracing.ChildOf(spCtx))
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

func startSQSExitSpan(op string, req *request.Request, sensor *instana.Sensor) {
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
		opentracing.Tags{
			"sqs.sort": "exit",
			"sqs.type": op,
		},
		tags,
	)

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
	injectTraceContext(sp, req)
}
