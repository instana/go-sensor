// +build go1.13

// In the upcoming v1.8.1 github.com/instana/go-sensor will start sending GRPC spans as registered spans
// that sends RPC tags within the `data.rpc` section instead of `data.sdk`. This is an interim test suite
// that assumes go-sensor@v1.8.0. It will be removed once v1.8.1 is released and the version in go.mod is
// updated.

package instagrpc_test

import (
	"encoding/json"
	"fmt"

	ot "github.com/opentracing/opentracing-go"
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
	Kind  int    `json:"k"`
	Error bool   `json:"error"`
	Ec    int    `json:"ec,omitempty"`
	Lang  string `json:"ta,omitempty"`
	Data  struct {
		Service string `json:"service"`
		SDK     struct {
			Name      string `json:"name"`
			Type      string `json:"type"`
			Arguments string `json:"arguments"`
			Return    string `json:"return"`
			Custom    struct {
				Tags    ot.Tags                           `json:"tags"`
				Logs    map[uint64]map[string]interface{} `json:"logs"`
				Baggage map[string]string                 `json:"baggage"`
			} `json:"custom"`
		} `json:"sdk"`
	} `json:"data"`
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
