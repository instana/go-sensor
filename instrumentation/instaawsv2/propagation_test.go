// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: This file contains the testcases for testing the private methods of the package.

func TestSpanContextFromSQSMessage(t *testing.T) {
	opts := instana.DefaultOptions()
	opts.Recorder = instana.NewTestRecorder()
	c := instana.InitCollector(opts)
	defer instana.ShutdownCollector()

	examples := map[string]sqstypes.Message{
		"standard keys": {
			MessageAttributes: map[string]sqstypes.MessageAttributeValue{
				"Custom": {
					DataType:    stringRef("String"),
					StringValue: stringRef("custom attribute"),
				},
				"X_INSTANA_T": {
					DataType:    stringRef("String"),
					StringValue: stringRef("00000000000000010000000000000002"),
				},
				"X_INSTANA_S": {
					DataType:    stringRef("String"),
					StringValue: stringRef("0000000000000003"),
				},
				"X_INSTANA_L": {
					DataType:    stringRef("String"),
					StringValue: stringRef("1"),
				},
			},
		},
	}

	for name, msg := range examples {
		t.Run(name, func(t *testing.T) {
			spCtx, err := SpanContextFromSQSMessage(msg, c)
			assert.NoError(t, err)
			assert.Equal(t, instana.SpanContext{
				TraceIDHi: 0x01,
				TraceID:   0x02,
				SpanID:    0x03,
				Baggage:   make(map[string]string),
			}, spCtx)
		})
	}

	t.Run("no context", func(t *testing.T) {
		_, err := SpanContextFromSQSMessage(sqstypes.Message{}, c)
		assert.Error(t, err)
	})
}

func TestSQSMessageAttributesCarrier_Set_FieldT(t *testing.T) {
	attrs := make(map[string]sqstypes.MessageAttributeValue)
	c := sqsMessageAttributesCarrier(attrs)

	c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")
	assert.Equal(t, map[string]sqstypes.MessageAttributeValue{
		fieldT: {
			DataType:    aws.String("String"),
			StringValue: aws.String("0000000000000001deadbeefdeadbeef"),
		},
	}, attrs)
}

func TestSQSMessageAttributesCarrier_Update_FieldT(t *testing.T) {
	examples := map[string]map[string]sqstypes.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_T": {
				DataType:    aws.String("String"),
				StringValue: aws.String("0000000000000000abcdef12abcdef12"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := sqsMessageAttributesCarrier(attrs)

			c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")
			assert.Equal(t, map[string]sqstypes.MessageAttributeValue{
				fieldT: {
					DataType:    stringRef("String"),
					StringValue: stringRef("0000000000000001deadbeefdeadbeef"),
				},
			}, attrs)
		})
	}
}

func TestSQSMessageAttributesCarrier_Set_FieldS(t *testing.T) {
	attrs := make(map[string]sqstypes.MessageAttributeValue)
	c := sqsMessageAttributesCarrier(attrs)

	c.Set(instana.FieldS, "deadbeefdeadbeef")
	assert.Equal(t, map[string]sqstypes.MessageAttributeValue{
		fieldS: {
			DataType:    stringRef("String"),
			StringValue: stringRef("deadbeefdeadbeef"),
		},
	}, attrs)
}

func TestSQSMessageAttributesCarrier_Update_FieldS(t *testing.T) {
	examples := map[string]map[string]sqstypes.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_S": {
				DataType:    stringRef("String"),
				StringValue: stringRef("abcdef12abcdef12"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := sqsMessageAttributesCarrier(attrs)

			c.Set(instana.FieldS, "deadbeefdeadbeef")
			assert.Equal(t, map[string]sqstypes.MessageAttributeValue{
				fieldS: {
					DataType:    stringRef("String"),
					StringValue: stringRef("deadbeefdeadbeef"),
				},
			}, attrs)
		})
	}
}

func TestSQSMessageAttributesCarrier_Set_FieldL(t *testing.T) {
	attrs := make(map[string]sqstypes.MessageAttributeValue)
	c := sqsMessageAttributesCarrier(attrs)

	c.Set(instana.FieldL, "1")
	assert.Equal(t, map[string]sqstypes.MessageAttributeValue{
		fieldL: {
			DataType:    stringRef("String"),
			StringValue: stringRef("1"),
		},
	}, attrs)
}

func TestSQSMessageAttributesCarrier_Update_FieldL(t *testing.T) {
	examples := map[string]map[string]sqstypes.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_L": {
				DataType:    stringRef("String"),
				StringValue: stringRef("1"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := sqsMessageAttributesCarrier(attrs)

			c.Set(instana.FieldL, "0")
			assert.Equal(t, map[string]sqstypes.MessageAttributeValue{
				fieldL: {
					DataType:    stringRef("String"),
					StringValue: stringRef("0"),
				},
			}, attrs)
		})
	}
}

func TestSQSMessageAttributesCarrier_ForeachKey(t *testing.T) {
	examples := map[string]map[string]sqstypes.MessageAttributeValue{
		"standard keys": {
			"Custom": {
				DataType:    stringRef("String"),
				StringValue: stringRef("custom attribute"),
			},
			"X_INSTANA_T": {
				DataType:    stringRef("String"),
				StringValue: stringRef("0000000000000001deadbeefdeadbeef"),
			},
			"X_INSTANA_S": {
				DataType:    stringRef("String"),
				StringValue: stringRef("abcdef12abcdef12"),
			},
			"X_INSTANA_L": {
				DataType:    stringRef("String"),
				StringValue: stringRef("1"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := sqsMessageAttributesCarrier(attrs)

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
	attrs := make(map[string]snstypes.MessageAttributeValue)
	c := snsMessageAttributesCarrier(attrs)

	c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")
	assert.Equal(t, map[string]snstypes.MessageAttributeValue{
		fieldT: {
			DataType:    stringRef("String"),
			StringValue: stringRef("0000000000000001deadbeefdeadbeef"),
		},
	}, attrs)
}

func TestSNSMessageAttributesCarrier_Update_FieldT(t *testing.T) {
	examples := map[string]map[string]snstypes.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_T": {
				DataType:    stringRef("String"),
				StringValue: stringRef("0000000000000000abcdef12abcdef12"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := snsMessageAttributesCarrier(attrs)

			c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")
			assert.Equal(t, map[string]snstypes.MessageAttributeValue{
				fieldT: {
					DataType:    stringRef("String"),
					StringValue: stringRef("0000000000000001deadbeefdeadbeef"),
				},
			}, attrs)
		})
	}
}

func TestSNSMessageAttributesCarrier_Set_FieldS(t *testing.T) {
	attrs := make(map[string]snstypes.MessageAttributeValue)
	c := snsMessageAttributesCarrier(attrs)

	c.Set(instana.FieldS, "deadbeefdeadbeef")
	assert.Equal(t, map[string]snstypes.MessageAttributeValue{
		fieldS: {
			DataType:    stringRef("String"),
			StringValue: stringRef("deadbeefdeadbeef"),
		},
	}, attrs)
}

func TestSNSMessageAttributesCarrier_Update_FieldS(t *testing.T) {
	examples := map[string]map[string]snstypes.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_S": {
				DataType:    stringRef("String"),
				StringValue: stringRef("abcdef12abcdef12"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := snsMessageAttributesCarrier(attrs)

			c.Set(instana.FieldS, "deadbeefdeadbeef")
			assert.Equal(t, map[string]snstypes.MessageAttributeValue{
				fieldS: {
					DataType:    stringRef("String"),
					StringValue: stringRef("deadbeefdeadbeef"),
				},
			}, attrs)
		})
	}
}

func TestSNSMessageAttributesCarrier_Set_FieldL(t *testing.T) {
	attrs := make(map[string]snstypes.MessageAttributeValue)
	c := snsMessageAttributesCarrier(attrs)

	c.Set(instana.FieldL, "1")
	assert.Equal(t, map[string]snstypes.MessageAttributeValue{
		fieldL: {
			DataType:    stringRef("String"),
			StringValue: stringRef("1"),
		},
	}, attrs)
}

func TestSNSMessageAttributesCarrier_Update_FieldL(t *testing.T) {
	examples := map[string]map[string]snstypes.MessageAttributeValue{
		"standard key": {
			"X_INSTANA_L": {
				DataType:    stringRef("String"),
				StringValue: stringRef("1"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := snsMessageAttributesCarrier(attrs)

			c.Set(instana.FieldL, "0")
			assert.Equal(t, map[string]snstypes.MessageAttributeValue{
				fieldL: {
					DataType:    stringRef("String"),
					StringValue: stringRef("0"),
				},
			}, attrs)
		})
	}
}

func TestSNSMessageAttributesCarrier_ForeachKey(t *testing.T) {
	examples := map[string]map[string]snstypes.MessageAttributeValue{
		"standard keys": {
			"Custom": {
				DataType:    stringRef("String"),
				StringValue: stringRef("custom attribute"),
			},
			"X_INSTANA_T": {
				DataType:    stringRef("String"),
				StringValue: stringRef("0000000000000001deadbeefdeadbeef"),
			},
			"X_INSTANA_S": {
				DataType:    stringRef("String"),
				StringValue: stringRef("abcdef12abcdef12"),
			},
			"X_INSTANA_L": {
				DataType:    stringRef("String"),
				StringValue: stringRef("1"),
			},
		},
	}

	for name, attrs := range examples {
		t.Run(name, func(t *testing.T) {
			c := snsMessageAttributesCarrier(attrs)

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
