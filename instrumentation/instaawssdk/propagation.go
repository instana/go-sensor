// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
)

const (
	// FieldT is the trace ID message attribute key
	FieldT = "X_INSTANA_T"
	// FieldS is the span ID message attribute key
	FieldS = "X_INSTANA_S"
	// FieldL is the trace level message attribute key
	FieldL = "X_INSTANA_L"
)

// SpanContextFromSQSMessage returns the trace context from an SQS message
func SpanContextFromSQSMessage(msg *sqs.Message, sensor *instana.Sensor) (opentracing.SpanContext, bool) {
	spanContext, err := sensor.Tracer().Extract(
		opentracing.TextMap,
		SQSMessageAttributesCarrier(msg.MessageAttributes),
	)
	if err != nil {
		return nil, false
	}

	return spanContext, true
}

// SQSMessageAttributesCarrier creates a new trace context carrier suitable for (opentracing.Tracer).Inject()
// that uses SQS message attributes as a storage
func SQSMessageAttributesCarrier(attrs map[string]*sqs.MessageAttributeValue) messageAttributesCarrier {
	return messageAttributesCarrier{
		Attrs: sqsMessageAttributes(attrs),
	}
}

// SNSMessageAttributesCarrier creates a new trace context carrier suitable for (opentracing.Tracer).Inject()
// that uses SNS message attributes as a storage
func SNSMessageAttributesCarrier(attrs map[string]*sns.MessageAttributeValue) messageAttributesCarrier {
	return messageAttributesCarrier{
		Attrs: snsMessageAttributes(attrs),
	}
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
		c.Attrs.Set(FieldT, val)
	case instana.FieldS:
		c.Attrs.Set(FieldS, val)
	case instana.FieldL:
		c.Attrs.Set(FieldL, val)
	}
}

func (c messageAttributesCarrier) ForeachKey(handler func(key, val string) error) error {
	if v, ok := c.Attrs.Get(FieldT); ok {
		handler(instana.FieldT, v)
	}

	if v, ok := c.Attrs.Get(FieldS); ok {
		handler(instana.FieldS, v)
	}

	if v, ok := c.Attrs.Get(FieldL); ok {
		handler(instana.FieldL, v)
	}

	return nil
}

type sqsMessageAttributes map[string]*sqs.MessageAttributeValue

func (attrs sqsMessageAttributes) Get(key string) (string, bool) {
	if v, ok := attrs[key]; ok {
		return aws.StringValue(v.StringValue), ok
	}

	return "", false
}

func (attrs sqsMessageAttributes) Set(key, val string) {
	attrs[key] = &sqs.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(val),
	}
}

func (attrs sqsMessageAttributes) Del(key string) {
	delete(attrs, key)
}

type snsMessageAttributes map[string]*sns.MessageAttributeValue

func (attrs snsMessageAttributes) Get(key string) (string, bool) {
	if v, ok := attrs[key]; ok {
		return aws.StringValue(v.StringValue), ok
	}

	return "", false
}

func (attrs snsMessageAttributes) Set(key, val string) {
	attrs[key] = &sns.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(val),
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
			(*attrs)[k] = &sns.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(v.Value),
			}
		case "binary":
			val, err := base64.StdEncoding.DecodeString(v.Value)
			if err == nil {
				(*attrs)[k] = &sns.MessageAttributeValue{
					DataType:    aws.String("Binary"),
					BinaryValue: val,
				}
			}
		default:
			// skip
		}
	}

	return nil
}
