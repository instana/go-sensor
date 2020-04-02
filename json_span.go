package instana

// Span represents the OpenTracing span document to be sent to the agent
type Span struct {
	TraceID   int64       `json:"t"`
	ParentID  int64      `json:"p,omitempty"`
	SpanID    int64       `json:"s"`
	Timestamp uint64      `json:"ts"`
	Duration  uint64      `json:"d"`
	Name      string      `json:"n"`
	From      *fromS      `json:"f"`
	Kind      int         `json:"k"`
	Error     bool        `json:"error"`
	Ec        int         `json:"ec,omitempty"`
	Lang      string      `json:"ta,omitempty"`
	Data      SDKSpanData `json:"data"`
}

// SpanData contains fields to be sent in the `data` section of an OT span document. These fields are
// common for all span types.
type SpanData struct {
	Service string `json:"service,omitempty"`
}

// SDKSpanData represents the `data` section of an SDK span sent within an OT span document
type SDKSpanData struct {
	SpanData
	Tags SDKSpanTags `json:"sdk"`
}

// SDKSpanTags contains fields within the `data.sdk` section of an OT span document
type SDKSpanTags struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type,omitempty"`
	Arguments string                 `json:"arguments,omitempty"`
	Return    string                 `json:"return,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}
