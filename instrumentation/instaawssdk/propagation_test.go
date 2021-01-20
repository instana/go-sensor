// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

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
		"legacy key": {
			"X_INSTANA_ST": {
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
		"legacy key": {
			"X_INSTANA_SS": {
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
		"legacy key": {
			"X_INSTANA_SL": {
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
		"legacy keys": {
			"Custom": {
				DataType:    aws.String("String"),
				StringValue: aws.String("custom attribute"),
			},
			"X_INSTANA_ST": {
				DataType:    aws.String("String"),
				StringValue: aws.String("0000000000000001deadbeefdeadbeef"),
			},
			"X_INSTANA_SS": {
				DataType:    aws.String("String"),
				StringValue: aws.String("abcdef12abcdef12"),
			},
			"X_INSTANA_SL": {
				DataType:    aws.String("String"),
				StringValue: aws.String("1"),
			},
		},
		"legacy and standard keys": {
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
			"X_INSTANA_ST": {
				DataType:    aws.String("String"),
				StringValue: aws.String("00000000000000001212121212121212"),
			},
			"X_INSTANA_SS": {
				DataType:    aws.String("String"),
				StringValue: aws.String("2323232323232323"),
			},
			"X_INSTANA_SL": {
				DataType:    aws.String("String"),
				StringValue: aws.String("0"),
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
