// +build go1.9

package instasarama_test

import (
	"encoding/json"
	"fmt"
)

type agentSpan struct {
	TraceID   string `json:"t"`
	ParentID  string `json:"p,omitempty"`
	SpanID    string `json:"s"`
	Timestamp uint64 `json:"ts"`
	Duration  uint64 `json:"d"`
	Name      string `json:"n"`
	From      struct {
		PID    string `json:"e"`
		HostID string `json:"h"`
	} `json:"f"`
	Batch struct {
		Size int `json:"s"`
	} `json:"b"`
	Kind int `json:"k"`
	Ec   int `json:"ec,omitempty"`
	Data struct {
		Service string             `json:"service"`
		Kafka   agentKafkaSpanData `json:"kafka"`
	} `json:"data"`
}

type agentKafkaSpanData struct {
	Service string `json:"service"`
	Access  string `json:"access"`
}

// unmarshalAgentSpan is a helper function that copies span data values
// into an agentSpan to not to depend on the implementation of instana.Recorder
func extractAgentSpan(span interface{}) (agentSpan, error) {
	d, err := json.Marshal(span)
	if err != nil {
		return agentSpan{}, fmt.Errorf("failed to marshal agent span data: %s", err)
	}

	var data agentSpan
	if err := json.Unmarshal(d, &data); err != nil {
		return agentSpan{}, fmt.Errorf("failed to unmarshal agent span data: %s", err)
	}

	return data, nil
}
