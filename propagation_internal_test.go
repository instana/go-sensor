// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLevel(t *testing.T) {
	examples := map[string]struct {
		Value               string
		ExpectedSuppressed  bool
		ExpectedCorrelation EUMCorrelationData
	}{
		"empty header": {},
		"level=0, no correlation id": {
			Value:              "0",
			ExpectedSuppressed: true,
		},
		"level=1, no correlation id": {
			Value: "1",
		},
		"level=0, with correlation id": {
			Value:              "0   ,   correlationType=web  ;\t  correlationId=Test Value",
			ExpectedSuppressed: true,
			ExpectedCorrelation: EUMCorrelationData{
				Type: "web",
				ID:   "Test Value",
			},
		},
		"level=1, with correlation id": {
			Value: "1,correlationType=mobile;correlationId=Test Value",
			ExpectedCorrelation: EUMCorrelationData{
				Type: "mobile",
				ID:   "Test Value",
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			suppressed, corrData, err := parseLevel(example.Value)
			require.NoError(t, err)
			assert.Equal(t, example.ExpectedSuppressed, suppressed)
			assert.Equal(t, example.ExpectedCorrelation, corrData)
		})
	}
}
