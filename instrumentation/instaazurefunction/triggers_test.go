// (c) Copyright IBM Corp. 2023

package instaazurefunction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPayloadParsing(t *testing.T) {
	testcases := map[string]struct {
		Payload  string
		Expected Metadata
	}{
		"success": {
			Payload: `{"Metadata":{"Headers":{"User-Agent":"curl/7.79.1"},"sys":{"MethodName":"roboshop"}}}`,
			Expected: Metadata{Headers: struct {
				UserAgent string `json:"User-Agent"`
			}{
				UserAgent: "curl/7.79.1",
			}, Sys: struct {
				MethodName string `json:"MethodName"`
			}{
				MethodName: "roboshop",
			}},
		},
		"empty json": {
			Payload:  `{}`,
			Expected: Metadata{},
		},
	}

	for name, test := range testcases {
		t.Run(name, func(t *testing.T) {
			Actual, err := extractSpanData([]byte(test.Payload))

			assert.NoError(t, err)
			assert.Equal(t, Actual, test.Expected)
		})
	}
}

func TestMetaDataAPIs(t *testing.T) {
	type Tags struct {
		TriggerType string
		MethodName  string
	}

	testcases := map[string]struct {
		Payload  string
		Expected Tags
	}{
		"http": {
			Payload: `{"Metadata":{"Headers":{"User-Agent":"curl/7.79.1"},"sys":{"MethodName":"roboshop"}}}`,
			Expected: Tags{
				TriggerType: "HTTP",
				MethodName:  "roboshop",
			},
		},
		"queue": {
			Payload: `{"Metadata":{"InsertionTime":"random time value","sys":{"MethodName":"roboshop"}}}`,
			Expected: Tags{
				TriggerType: "Queue",
				MethodName:  "roboshop",
			},
		},
	}

	for name, test := range testcases {
		t.Run(name, func(t *testing.T) {
			metaData, err := extractSpanData([]byte(test.Payload))

			assert.NoError(t, err)
			assert.Equal(t, test.Expected.MethodName, metaData.functionName())
			assert.Equal(t, test.Expected.TriggerType, metaData.triggerName())
		})
	}
}

func TestErrorJson(t *testing.T) {
	payload := ""
	metaData, err := extractSpanData([]byte(payload))

	assert.Error(t, err)
	assert.Equal(t, Metadata{}, metaData)
}
