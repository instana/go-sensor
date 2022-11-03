// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpanContextFromSQSMessage(t *testing.T) {
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(
			instana.DefaultOptions(),
			instana.NewTestRecorder(),
		),
	)
	defer instana.ShutdownSensor()

	examples := map[string]*sqs.Message{
		"standard keys": {
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"Custom": {
					DataType:    aws.String("String"),
					StringValue: aws.String("custom attribute"),
				},
				"X_INSTANA_T": {
					DataType:    aws.String("String"),
					StringValue: aws.String("00000000000000010000000000000002"),
				},
				"X_INSTANA_S": {
					DataType:    aws.String("String"),
					StringValue: aws.String("0000000000000003"),
				},
				"X_INSTANA_L": {
					DataType:    aws.String("String"),
					StringValue: aws.String("1"),
				},
			},
		},
	}

	for name, msg := range examples {
		t.Run(name, func(t *testing.T) {
			spCtx, ok := instaawssdk.SpanContextFromSQSMessage(msg, sensor)
			require.True(t, ok)
			assert.Equal(t, instana.SpanContext{
				TraceIDHi: 0x01,
				TraceID:   0x02,
				SpanID:    0x03,
				Baggage:   make(map[string]string),
			}, spCtx)
		})
	}

	t.Run("no context", func(t *testing.T) {
		_, ok := instaawssdk.SpanContextFromSQSMessage(&sqs.Message{}, sensor)
		assert.False(t, ok)
	})
}

func TestSQSMessageAttributesCarrier_Set_FieldT(t *testing.T) {
	attrs := make(map[string]*sqs.MessageAttributeValue)
	c := instaawssdk.SQSMessageAttributesCarrier(attrs)

	c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")
	assert.Equal(t, map[string]*sqs.MessageAttributeValue{
		instaawssdk.FieldT: {
			DataType:    aws.String("String"),
			StringValue: aws.String("0000000000000001deadbeefdeadbeef"),
		},
	}, attrs)
}

func TestSQSMessageAttributesCarrier_Update_FieldT(t *testing.T) {
	examples := map[string]map[string]*sqs.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_T": {
				DataType:    aws.String("String"),
				StringValue: aws.String("0000000000000000abcdef12abcdef12"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := instaawssdk.SQSMessageAttributesCarrier(attrs)

			c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")
			assert.Equal(t, map[string]*sqs.MessageAttributeValue{
				instaawssdk.FieldT: {
					DataType:    aws.String("String"),
					StringValue: aws.String("0000000000000001deadbeefdeadbeef"),
				},
			}, attrs)
		})
	}
}

func TestSQSMessageAttributesCarrier_Set_FieldS(t *testing.T) {
	attrs := make(map[string]*sqs.MessageAttributeValue)
	c := instaawssdk.SQSMessageAttributesCarrier(attrs)

	c.Set(instana.FieldS, "deadbeefdeadbeef")
	assert.Equal(t, map[string]*sqs.MessageAttributeValue{
		instaawssdk.FieldS: {
			DataType:    aws.String("String"),
			StringValue: aws.String("deadbeefdeadbeef"),
		},
	}, attrs)
}

func TestSQSMessageAttributesCarrier_Update_FieldS(t *testing.T) {
	examples := map[string]map[string]*sqs.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_S": {
				DataType:    aws.String("String"),
				StringValue: aws.String("abcdef12abcdef12"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := instaawssdk.SQSMessageAttributesCarrier(attrs)

			c.Set(instana.FieldS, "deadbeefdeadbeef")
			assert.Equal(t, map[string]*sqs.MessageAttributeValue{
				instaawssdk.FieldS: {
					DataType:    aws.String("String"),
					StringValue: aws.String("deadbeefdeadbeef"),
				},
			}, attrs)
		})
	}
}

func TestSQSMessageAttributesCarrier_Set_FieldL(t *testing.T) {
	attrs := make(map[string]*sqs.MessageAttributeValue)
	c := instaawssdk.SQSMessageAttributesCarrier(attrs)

	c.Set(instana.FieldL, "1")
	assert.Equal(t, map[string]*sqs.MessageAttributeValue{
		instaawssdk.FieldL: {
			DataType:    aws.String("String"),
			StringValue: aws.String("1"),
		},
	}, attrs)
}

func TestSQSMessageAttributesCarrier_Update_FieldL(t *testing.T) {
	examples := map[string]map[string]*sqs.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_L": {
				DataType:    aws.String("String"),
				StringValue: aws.String("1"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := instaawssdk.SQSMessageAttributesCarrier(attrs)

			c.Set(instana.FieldL, "0")
			assert.Equal(t, map[string]*sqs.MessageAttributeValue{
				instaawssdk.FieldL: {
					DataType:    aws.String("String"),
					StringValue: aws.String("0"),
				},
			}, attrs)
		})
	}
}

func TestSQSMessageAttributesCarrier_ForeachKey(t *testing.T) {
	examples := map[string]map[string]*sqs.MessageAttributeValue{
		"standard keys": {
			"Custom": {
				DataType:    aws.String("String"),
				StringValue: aws.String("custom attribute"),
			},
			"X_INSTANA_T": {
				DataType:    aws.String("String"),
				StringValue: aws.String("0000000000000001deadbeefdeadbeef"),
			},
			"X_INSTANA_S": {
				DataType:    aws.String("String"),
				StringValue: aws.String("abcdef12abcdef12"),
			},
			"X_INSTANA_L": {
				DataType:    aws.String("String"),
				StringValue: aws.String("1"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := instaawssdk.SQSMessageAttributesCarrier(attrs)

			collected := make(map[string]string)
			require.NoError(t, c.ForeachKey(func(key, val string) error {
				if v, ok := collected[key]; ok {
					return fmt.Errorf("duplicate key %q (previous value was %q)", key, v)
				}

				collected[key] = val

				return nil
			}))

			assert.Equal(t, map[string]string{
				instana.FieldT: "0000000000000001deadbeefdeadbeef",
				instana.FieldS: "abcdef12abcdef12",
				instana.FieldL: "1",
			}, collected)
		})
	}
}

func TestSNSMessageAttributesCarrier_Set_FieldT(t *testing.T) {
	attrs := make(map[string]*sns.MessageAttributeValue)
	c := instaawssdk.SNSMessageAttributesCarrier(attrs)

	c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")
	assert.Equal(t, map[string]*sns.MessageAttributeValue{
		instaawssdk.FieldT: {
			DataType:    aws.String("String"),
			StringValue: aws.String("0000000000000001deadbeefdeadbeef"),
		},
	}, attrs)
}

func TestSNSMessageAttributesCarrier_Update_FieldT(t *testing.T) {
	examples := map[string]map[string]*sns.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_T": {
				DataType:    aws.String("String"),
				StringValue: aws.String("0000000000000000abcdef12abcdef12"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := instaawssdk.SNSMessageAttributesCarrier(attrs)

			c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")
			assert.Equal(t, map[string]*sns.MessageAttributeValue{
				instaawssdk.FieldT: {
					DataType:    aws.String("String"),
					StringValue: aws.String("0000000000000001deadbeefdeadbeef"),
				},
			}, attrs)
		})
	}
}

func TestSNSMessageAttributesCarrier_Set_FieldS(t *testing.T) {
	attrs := make(map[string]*sns.MessageAttributeValue)
	c := instaawssdk.SNSMessageAttributesCarrier(attrs)

	c.Set(instana.FieldS, "deadbeefdeadbeef")
	assert.Equal(t, map[string]*sns.MessageAttributeValue{
		instaawssdk.FieldS: {
			DataType:    aws.String("String"),
			StringValue: aws.String("deadbeefdeadbeef"),
		},
	}, attrs)
}

func TestSNSMessageAttributesCarrier_Update_FieldS(t *testing.T) {
	examples := map[string]map[string]*sns.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_S": {
				DataType:    aws.String("String"),
				StringValue: aws.String("abcdef12abcdef12"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := instaawssdk.SNSMessageAttributesCarrier(attrs)

			c.Set(instana.FieldS, "deadbeefdeadbeef")
			assert.Equal(t, map[string]*sns.MessageAttributeValue{
				instaawssdk.FieldS: {
					DataType:    aws.String("String"),
					StringValue: aws.String("deadbeefdeadbeef"),
				},
			}, attrs)
		})
	}
}

func TestSNSMessageAttributesCarrier_Set_FieldL(t *testing.T) {
	attrs := make(map[string]*sns.MessageAttributeValue)
	c := instaawssdk.SNSMessageAttributesCarrier(attrs)

	c.Set(instana.FieldL, "1")
	assert.Equal(t, map[string]*sns.MessageAttributeValue{
		instaawssdk.FieldL: {
			DataType:    aws.String("String"),
			StringValue: aws.String("1"),
		},
	}, attrs)
}

func TestSNSMessageAttributesCarrier_Update_FieldL(t *testing.T) {
	examples := map[string]map[string]*sns.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_L": {
				DataType:    aws.String("String"),
				StringValue: aws.String("1"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := instaawssdk.SNSMessageAttributesCarrier(attrs)

			c.Set(instana.FieldL, "0")
			assert.Equal(t, map[string]*sns.MessageAttributeValue{
				instaawssdk.FieldL: {
					DataType:    aws.String("String"),
					StringValue: aws.String("0"),
				},
			}, attrs)
		})
	}
}

func TestSNSMessageAttributesCarrier_ForeachKey(t *testing.T) {
	examples := map[string]map[string]*sns.MessageAttributeValue{
		"standard keys": {
			"Custom": {
				DataType:    aws.String("String"),
				StringValue: aws.String("custom attribute"),
			},
			"X_INSTANA_T": {
				DataType:    aws.String("String"),
				StringValue: aws.String("0000000000000001deadbeefdeadbeef"),
			},
			"X_INSTANA_S": {
				DataType:    aws.String("String"),
				StringValue: aws.String("abcdef12abcdef12"),
			},
			"X_INSTANA_L": {
				DataType:    aws.String("String"),
				StringValue: aws.String("1"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := instaawssdk.SNSMessageAttributesCarrier(attrs)

			collected := make(map[string]string)
			require.NoError(t, c.ForeachKey(func(key, val string) error {
				if v, ok := collected[key]; ok {
					return fmt.Errorf("duplicate key %q (previous value was %q)", key, v)
				}

				collected[key] = val

				return nil
			}))

			assert.Equal(t, map[string]string{
				instana.FieldT: "0000000000000001deadbeefdeadbeef",
				instana.FieldS: "abcdef12abcdef12",
				instana.FieldL: "1",
			}, collected)
		})
	}
}
