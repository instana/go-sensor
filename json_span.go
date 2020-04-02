package instana

type typedSpanData interface {
	Type() RegisteredSpanType
}

// Registered types supported by Instana. The span type is determined based on
// the operation name passed to the `StartSpan()` call of a tracer.
//
// It is NOT RECOMMENDED to use operation names that match any of these constants in your
// custom instrumentation  code unless you explicitly wish to send data as a registered span.
// The conversion will result in loss of custom tags that are not supported for this span type.
// The list of supported tags can be found in the godoc of the respective span tags type below.
const (
	// SDK span, a generic span containing arbitrary data. Spans with operation name
	// not listed in the subsequent list will be sent as an SDK spans forwarding all
	// attached tags to the agent
	SDKSpanType = RegisteredSpanType("sdk")
)

// RegisteredSpanType represents the span type supported by Instana
type RegisteredSpanType string

// ExtractData is a factory method to create the `data` section for a typed span
func (st RegisteredSpanType) ExtractData(span *spanS) typedSpanData {
	switch st {
	default:
		return NewSDKSpanData(span)
	}
}

// Span represents the OpenTracing span document to be sent to the agent
type Span struct {
	TraceID   int64         `json:"t"`
	ParentID  int64         `json:"p,omitempty"`
	SpanID    int64         `json:"s"`
	Timestamp uint64        `json:"ts"`
	Duration  uint64        `json:"d"`
	Name      string        `json:"n"`
	From      *fromS        `json:"f"`
	Kind      int           `json:"k"`
	Error     bool          `json:"error"`
	Ec        int           `json:"ec,omitempty"`
	Lang      string        `json:"ta,omitempty"`
	Data      typedSpanData `json:"data"`
}

// SpanData contains fields to be sent in the `data` section of an OT span document. These fields are
// common for all span types.
type SpanData struct {
	Service string `json:"service,omitempty"`
	st      RegisteredSpanType
}

// NewSpanData initializes a new span data from tracer span
func NewSpanData(span *spanS, st RegisteredSpanType) SpanData {
	return SpanData{
		Service: span.Service,
		st:      st,
	}
}

// Name returns the registered name for the span type suitable for use as the value of `n` field.
func (d SpanData) Type() RegisteredSpanType {
	return d.st
}

// SDKSpanData represents the `data` section of an SDK span sent within an OT span document
type SDKSpanData struct {
	SpanData
	Tags SDKSpanTags `json:"sdk"`
}

// NewSDKSpanData initializes a new SDK span data from tracer span
func NewSDKSpanData(span *spanS) SDKSpanData {
	return SDKSpanData{
		SpanData: NewSpanData(span, SDKSpanType),
		Tags:     NewSDKSpanTags(span),
	}
}

// SDKSpanTags contains fields within the `data.sdk` section of an OT span document
type SDKSpanTags struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type,omitempty"`
	Arguments string                 `json:"arguments,omitempty"`
	Return    string                 `json:"return,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// NewSDKSpanTags extracts SDK span tags from a tracer span
func NewSDKSpanTags(span *spanS) SDKSpanTags {
	tags := SDKSpanTags{
		Name:   span.Operation,
		Type:   span.Kind().String(),
		Custom: map[string]interface{}{},
	}

	if len(span.Tags) != 0 {
		tags.Custom["tags"] = span.Tags
	}

	if logs := span.collectLogs(); len(logs) > 0 {
		tags.Custom["logs"] = logs
	}

	if len(span.context.Baggage) != 0 {
		tags.Custom["baggage"] = span.context.Baggage
	}

	return tags
}
