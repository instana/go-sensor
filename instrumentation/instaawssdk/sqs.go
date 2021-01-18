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

func startSQSEntrySpan(req *request.Request, sensor *instana.Sensor) {
	sp := sensor.Tracer().StartSpan("sqs",
		ext.SpanKindConsumer,
		opentracing.Tags{
			"sqs.sort": "entry",
		},
		extractSQSTags(req),
	)

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
}

func startSQSExitSpan(op string, req *request.Request, sensor *instana.Sensor) {
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
		extractSQSTags(req),
	)

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
}

// FinalizeSQSSpan retrieves tags from completed request.Request and adds them
// to the span
func FinalizeSQSSpan(sp opentracing.Span, req *request.Request) {
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

	if req.Error != nil {
		sp.SetTag("sqs.error", req.Error.Error())
	}
}
