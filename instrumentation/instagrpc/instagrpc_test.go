package instagrpc_test

import (
	"encoding/json"
	"fmt"
)

type agentSpan struct {
	TraceID   int64  `json:"t"`
	ParentID  int64  `json:"p,omitempty"`
	SpanID    int64  `json:"s"`
	Timestamp uint64 `json:"ts"`
	Duration  uint64 `json:"d"`
	Name      string `json:"n"`
	From      struct {
		PID    string `json:"e"`
		HostID string `json:"h"`
	} `json:"f"`
	Kind int `json:"k"`
	Ec   int `json:"ec,omitempty"`
	Data struct {
		Service string           `json:"service"`
		RPC     agentRPCSpanData `json:"rpc"`
	} `json:"data"`
}

type agentRPCSpanData struct {
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	Call     string `json:"call,omitempty"`
	CallType string `json:"call_type,omitempty"`
	Flavor   string `json:"flavor,omitempty"`
	Error    string `json:"error,omitempty"`
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
