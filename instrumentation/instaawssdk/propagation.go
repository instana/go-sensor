// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	instana "github.com/instana/go-sensor"
)

const (
	// FieldT is the trace ID message attribute key
	FieldT = "X_INSTANA_T"
	// FieldS is the span ID message attribute key
	FieldS = "X_INSTANA_S"
	// FieldL is the trace level message attribute key
	FieldL = "X_INSTANA_L"

	legacyFieldT = "X_INSTANA_ST"
	legacyFieldS = "X_INSTANA_SS"
	legacyFieldL = "X_INSTANA_SL"
)

// SQSMessageAttributesCarrier is a trace context carrier that propagates Instana
// OpenTracing headers via AWS SQS messages. It satisfies both opentracing.TextMapReader
// and opentracing.TextMapWriter and thus can be used for context propagation using an
// OpenTracing tracer.
type SQSMessageAttributesCarrier map[string]*sqs.MessageAttributeValue

// Set implements opentracing.TextMapWriter for SQSMessageAttributesCarrier
func (c SQSMessageAttributesCarrier) Set(key, val string) {
	switch strings.ToLower(key) {
	case instana.FieldT:
		if _, ok := c[legacyFieldT]; ok {
			delete(c, legacyFieldT)
		}

		c[FieldT] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(val),
		}
	case instana.FieldS:
		if _, ok := c[legacyFieldS]; ok {
			delete(c, legacyFieldS)
		}

		c[FieldS] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(val),
		}
	case instana.FieldL:
		if _, ok := c[legacyFieldL]; ok {
			delete(c, legacyFieldL)
		}

		c[FieldL] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(val),
		}
	}
}

// ForeachKey implements opentracing.TextMapReader for SQSMessageAttributesCarrier
func (c SQSMessageAttributesCarrier) ForeachKey(handler func(key, val string) error) error {
	if len(c) == 0 {
		return nil
	}

	if v, ok := c.getAttributeWithFallback(FieldT, legacyFieldT); ok {
		handler(instana.FieldT, aws.StringValue(v.StringValue))
	}

	if v, ok := c.getAttributeWithFallback(FieldS, legacyFieldS); ok {
		handler(instana.FieldS, aws.StringValue(v.StringValue))
	}

	if v, ok := c.getAttributeWithFallback(FieldL, legacyFieldL); ok {
		handler(instana.FieldL, aws.StringValue(v.StringValue))
	}

	return nil
}

func (c SQSMessageAttributesCarrier) getAttributeWithFallback(key, fallbackKey string) (*sqs.MessageAttributeValue, bool) {
	if v, ok := c[key]; ok {
		return v, ok
	}

	if v, ok := c[fallbackKey]; ok {
		return v, ok
	}

	return nil, false
}
