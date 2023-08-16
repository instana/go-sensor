// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	snsTypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// traceSQSMessage returns an entry span for an SQS message. The context of this span is injected
// into message attributes. This context can than be retrieved with instaawsv2.SpanContextFromSQSMessage()
// and used in the message handler method to continue the trace.
func traceSQSMessage(msg *sqsTypes.Message, tracer instana.TracerLogger) opentracing.Span {
	opts := []opentracing.StartSpanOption{
		ext.SpanKindConsumer,
		opentracing.Tags{
			sqsSort: "entry",
		},
	}

	if spCtx, ok := SpanContextFromSQSMessage(*msg, tracer); ok {
		opts = append(opts, opentracing.ChildOf(spCtx))
	} else {
		body := stringDeRef(msg.Body)

		// In case the delivery has been created via a subscription to an SNS topic,
		// the message body will be a JSON document containing the SNS notification
		// along with message attributes.
		var payload struct {
			MessageAttributes snsMessageAttributes `json:"MessageAttributes"`
		}

		// try to unmarshal the message attributes and extract the trace context from there
		if err := json.Unmarshal([]byte(body), &payload); err == nil {
			if spCtx, err := tracer.Tracer().Extract(
				opentracing.TextMap,
				snsMessageAttributesCarrier(payload.MessageAttributes),
			); err == nil {
				opts = append(opts, opentracing.ChildOf(spCtx))
			}
		}
	}

	sp := tracer.Tracer().StartSpan("sqs", opts...)

	if msg.MessageAttributes == nil {
		msg.MessageAttributes = make(map[string]sqsTypes.MessageAttributeValue)
	}

	sp.Tracer().Inject(
		sp.Context(),
		opentracing.TextMap,
		sqsMessageAttributesCarrier(msg.MessageAttributes),
	)

	return sp
}

// SpanContextFromSQSMessage returns the trace context from an SQS message
func SpanContextFromSQSMessage(msg sqsTypes.Message, tracer instana.TracerLogger) (opentracing.SpanContext, bool) {
	spanContext, err := tracer.Extract(
		opentracing.TextMap,
		sqsMessageAttributesCarrier(msg.MessageAttributes),
	)
	if err != nil {
		return nil, false
	}

	return spanContext, true
}

type sqsMessageAttributes map[string]sqsTypes.MessageAttributeValue

func (attrs sqsMessageAttributes) Get(key string) (string, bool) {
	if v, ok := attrs[key]; ok {
		return stringDeRef(v.StringValue), ok
	}

	return "", false
}

func (attrs sqsMessageAttributes) Set(key, val string) {
	attrs[key] = sqsTypes.MessageAttributeValue{
		DataType:    stringRef("String"),
		StringValue: stringRef(val),
	}
}

func (attrs sqsMessageAttributes) Del(key string) {
	delete(attrs, key)
}

type snsMessageAttributes map[string]snsTypes.MessageAttributeValue

func (attrs snsMessageAttributes) Get(key string) (string, bool) {
	if v, ok := attrs[key]; ok {
		return stringDeRef(v.StringValue), ok
	}

	return "", false
}

func (attrs snsMessageAttributes) Set(key, val string) {
	attrs[key] = snsTypes.MessageAttributeValue{
		DataType:    stringRef("String"),
		StringValue: stringRef(val),
	}
}

func (attrs snsMessageAttributes) Del(key string) {
	delete(attrs, key)
}

func (attrs *snsMessageAttributes) UnmarshalJSON(data []byte) error {
	var vs map[string]struct {
		Type  string `json:"Type"`
		Value string `json:"Value"`
	}

	if err := json.Unmarshal(data, &vs); err != nil {
		return err
	}

	if *attrs == nil {
		*attrs = make(snsMessageAttributes, len(vs))
	}

	for k, v := range vs {
		switch strings.ToLower(v.Type) {
		case "string":
			(*attrs)[k] = snsTypes.MessageAttributeValue{
				DataType:    stringRef("String"),
				StringValue: stringRef(v.Value),
			}
		case "binary":
			val, err := base64.StdEncoding.DecodeString(v.Value)
			if err == nil {
				(*attrs)[k] = snsTypes.MessageAttributeValue{
					DataType:    stringRef("Binary"),
					BinaryValue: val,
				}
			}
		default:
			// skip
		}
	}

	return nil
}

// snsMessageAttributesCarrier creates a new trace context carrier suitable for (opentracing.Tracer).Inject()
// that uses SNS message attributes as a storage
func snsMessageAttributesCarrier(attrs map[string]snsTypes.MessageAttributeValue) messageAttributesCarrier {
	return messageAttributesCarrier{
		Attrs: snsMessageAttributes(attrs),
	}
}

// sqsMessageAttributesCarrier creates a new trace context carrier suitable for (opentracing.Tracer).Inject()
// that uses SQS message attributes as a storage
func sqsMessageAttributesCarrier(attrs map[string]sqsTypes.MessageAttributeValue) messageAttributesCarrier {
	return messageAttributesCarrier{
		Attrs: sqsMessageAttributes(attrs),
	}
}
