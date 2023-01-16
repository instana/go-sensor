// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/ext"
)

type typedSpanData interface {
	Type() RegisteredSpanType
	Kind() SpanKind
}

// SpanKind represents values of field `k` in OpenTracing span representation. It represents
// the direction of the call associated with a span.
type SpanKind uint8

// Valid span kinds
const (
	// The kind of a span associated with an inbound call, this must be the first span in the trace.
	EntrySpanKind SpanKind = iota + 1
	// The kind of a span associated with an outbound call, e.g. an HTTP client request, posting to a message bus, etc.
	ExitSpanKind
	// The default kind for a span that is associated with a call within the same service.
	IntermediateSpanKind
)

// String returns string representation of a span kind suitable for use as a value for `data.sdk.type`
// tag of an SDK span. By default all spans are intermediate unless they are explicitly set to be "entry" or "exit"
func (k SpanKind) String() string {
	switch k {
	case EntrySpanKind:
		return "entry"
	case ExitSpanKind:
		return "exit"
	default:
		return "intermediate"
	}
}

// Span represents the OpenTracing span document to be sent to the agent
type Span struct {
	TraceID         int64
	TraceIDHi       int64
	ParentID        int64
	SpanID          int64
	Ancestor        *TraceReference
	Timestamp       uint64
	Duration        uint64
	Name            string
	From            *fromS
	Batch           *batchInfo
	Kind            int
	Ec              int
	Data            typedSpanData
	Synthetic       bool
	CorrelationType string
	CorrelationID   string
	ForeignTrace    bool
}

func newSpan(span *spanS) Span {
	data := RegisteredSpanType(span.Operation).extractData(span)
	sp := Span{
		TraceID:         span.context.TraceID,
		TraceIDHi:       span.context.TraceIDHi,
		ParentID:        span.context.ParentID,
		SpanID:          span.context.SpanID,
		Timestamp:       uint64(span.Start.UnixNano()) / uint64(time.Millisecond),
		Duration:        uint64(span.Duration) / uint64(time.Millisecond),
		Name:            string(data.Type()),
		Ec:              span.ErrorCount,
		CorrelationType: span.Correlation.Type,
		CorrelationID:   span.Correlation.ID,
		ForeignTrace:    span.context.ForeignTrace,
		Kind:            int(data.Kind()),
		Data:            data,
	}

	if bs, ok := span.Tags[batchSizeTag].(int); ok {
		if bs > 1 {
			sp.Batch = &batchInfo{Size: bs}
		}
		delete(span.Tags, batchSizeTag)
	}

	if syn, ok := span.Tags[syntheticCallTag].(bool); ok {
		sp.Synthetic = syn
		delete(span.Tags, syntheticCallTag)
	}

	if len(span.context.Links) > 0 {
		ancestor := span.context.Links[0]
		sp.Ancestor = &TraceReference{
			TraceID:  ancestor.TraceID,
			ParentID: ancestor.SpanID,
		}
	}

	return sp
}

// TraceReference is used to reference a parent span
type TraceReference struct {
	TraceID  string `json:"t"`
	ParentID string `json:"p,omitempty"`
}

// MarshalJSON serializes span to JSON for sending it to Instana
func (sp Span) MarshalJSON() ([]byte, error) {
	var parentID string
	if sp.ParentID != 0 {
		parentID = FormatID(sp.ParentID)
	}

	var longTraceID string
	if sp.TraceIDHi != 0 && sp.Kind == int(EntrySpanKind) {
		longTraceID = FormatLongID(sp.TraceIDHi, sp.TraceID)
	}

	return json.Marshal(struct {
		TraceReference

		SpanID          string          `json:"s"`
		LongTraceID     string          `json:"lt,omitempty"`
		Timestamp       uint64          `json:"ts"`
		Duration        uint64          `json:"d"`
		Name            string          `json:"n"`
		From            *fromS          `json:"f"`
		Batch           *batchInfo      `json:"b,omitempty"`
		Kind            int             `json:"k"`
		Ec              int             `json:"ec,omitempty"`
		Data            typedSpanData   `json:"data"`
		Synthetic       bool            `json:"sy,omitempty"`
		CorrelationType string          `json:"crtp,omitempty"`
		CorrelationID   string          `json:"crid,omitempty"`
		ForeignTrace    bool            `json:"tp,omitempty"`
		Ancestor        *TraceReference `json:"ia,omitempty"`
	}{
		TraceReference{
			FormatID(sp.TraceID),
			parentID,
		},
		FormatID(sp.SpanID),
		longTraceID,
		sp.Timestamp,
		sp.Duration,
		sp.Name,
		sp.From,
		sp.Batch,
		sp.Kind,
		sp.Ec,
		sp.Data,
		sp.Synthetic,
		sp.CorrelationType,
		sp.CorrelationID,
		sp.ForeignTrace,
		sp.Ancestor,
	})
}

type batchInfo struct {
	Size int `json:"s"`
}

// CustomSpanData holds user-defined span tags
type CustomSpanData struct {
	Tags map[string]interface{} `json:"tags,omitempty"`
}

func filterCustomSpanTags(tags map[string]interface{}, st RegisteredSpanType) map[string]interface{} {
	knownTags := st.TagsNames()
	customTags := make(map[string]interface{})

	for k, v := range tags {
		if k == string(ext.SpanKind) {
			continue
		}

		if _, ok := knownTags[k]; ok {
			continue
		}

		customTags[k] = v
	}

	return customTags
}

// SpanData contains fields to be sent in the `data` section of an OT span document. These fields are
// common for all span types.
type SpanData struct {
	Service string          `json:"service,omitempty"`
	Custom  *CustomSpanData `json:"sdk.custom,omitempty"`

	st RegisteredSpanType
}

// NewSpanData initializes a new span data from tracer span
func NewSpanData(span *spanS, st RegisteredSpanType) SpanData {
	data := SpanData{
		Service: span.Service,
		st:      st,
	}

	if customTags := filterCustomSpanTags(span.Tags, st); len(customTags) > 0 {
		data.Custom = &CustomSpanData{Tags: customTags}
	}

	return data
}

// Type returns the registered span type suitable for use as the value of `n` field.
func (d SpanData) Type() RegisteredSpanType {
	return d.st
}

// SDKSpanData represents the `data` section of an SDK span sent within an OT span document
type SDKSpanData struct {
	// Deprecated
	SpanData `json:"-"`

	Service string      `json:"service,omitempty"`
	Tags    SDKSpanTags `json:"sdk"`

	sk SpanKind
}

// newSDKSpanData initializes a new SDK span data from a tracer span
func newSDKSpanData(span *spanS) SDKSpanData {
	sk := IntermediateSpanKind

	switch span.Tags[string(ext.SpanKind)] {
	case ext.SpanKindRPCServerEnum, string(ext.SpanKindRPCServerEnum),
		ext.SpanKindConsumerEnum, string(ext.SpanKindConsumerEnum),
		"entry":
		sk = EntrySpanKind
	case ext.SpanKindRPCClientEnum, string(ext.SpanKindRPCClientEnum),
		ext.SpanKindProducerEnum, string(ext.SpanKindProducerEnum),
		"exit":
		sk = ExitSpanKind
	}

	return SDKSpanData{
		Service: span.Service,
		Tags:    newSDKSpanTags(span, sk.String()),
		sk:      sk,
	}
}

// Type returns the registered span type suitable for use as the value of `n` field.
func (d SDKSpanData) Type() RegisteredSpanType {
	return SDKSpanType
}

// Kind returns the kind of the span. It handles the github.com/opentracing/opentracing-go/ext.SpanKindEnum
// values as well as generic "entry" and "exit"
func (d SDKSpanData) Kind() SpanKind {
	return d.sk
}

// SDKSpanTags contains fields within the `data.sdk` section of an OT span document
type SDKSpanTags struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type,omitempty"`
	Arguments string                 `json:"arguments,omitempty"`
	Return    string                 `json:"return,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// newSDKSpanTags extracts SDK span tags from a tracer span
func newSDKSpanTags(span *spanS, spanType string) SDKSpanTags {
	tags := SDKSpanTags{
		Name:   span.Operation,
		Type:   spanType,
		Custom: map[string]interface{}{},
	}

	if len(span.Tags) != 0 {
		tags.Custom["tags"] = span.Tags
	}

	if len(span.context.Baggage) != 0 {
		tags.Custom["baggage"] = span.context.Baggage
	}

	return tags
}

// readStringTag populates the &dst with the tag value if it's of either string or []byte type
func readStringTag(dst *string, tag interface{}) {
	switch s := tag.(type) {
	case string:
		*dst = s
	case []byte:
		*dst = string(s)
	}
}

// readArrayStringTag populates &dst with the tag value if it's a slice of strings
func readArrayStringTag(dst *[]string, tag interface{}) {
	switch s := tag.(type) {
	case []string:
		*dst = s
	default:
		*dst = nil
	}
}

// readMapOfStringSlicesTag populates &dst with the tag value if it's a map of slice of strings
func readMapOfStringSlicesTag(dst *map[string][]string, tag interface{}) {
	if m, ok := tag.(map[string][]string); ok {
		*dst = m
	} else {
		*dst = nil
	}
}

// readBoolTag populates the &dst with the tag value if it's either bool, string, byte array or numeric type
func readBoolTag(dst *bool, tag interface{}) {
	switch s := tag.(type) {
	case string:
		val := strings.ToLower(s)
		if val == "true" || val == "1" {
			*dst = true
			return
		}
	case []byte:
		val := strings.ToLower(string(s))
		if val == "true" || val == "1" {
			*dst = true
			return
		}
	case bool:
		*dst = s
		return
	case uint:
		if s == 1 {
			*dst = true
			return
		}
	case uint8:
		if s == 1 {
			*dst = true
			return
		}
	case uint16:
		if s == 1 {
			*dst = true
			return
		}
	case uint32:
		if s == 1 {
			*dst = true
			return
		}
	case uint64:
		if s == 1 {
			*dst = true
			return
		}
	case int:
		if s == 1 {
			*dst = true
			return
		}
	case int8:
		if s == 1 {
			*dst = true
			return
		}
	case int16:
		if s == 1 {
			*dst = true
			return
		}
	case int32:
		if s == 1 {
			*dst = true
			return
		}
	case float32:
		if s == 1 {
			*dst = true
			return
		}
	case float64:
		if s == 1 {
			*dst = true
			return
		}
	}

	*dst = false
}

// readIntTag populates the &dst with the tag value if it's of any kind of integer type
func readIntTag(dst *int, tag interface{}) {
	switch n := tag.(type) {
	case int:
		*dst = n
	case int8:
		*dst = int(n)
	case int16:
		*dst = int(n)
	case int32:
		*dst = int(n)
	case int64:
		*dst = int(n)
	case uint:
		*dst = int(n)
	case uint8:
		*dst = int(n)
	case uint16:
		*dst = int(n)
	case uint32:
		*dst = int(n)
	case uint64:
		*dst = int(n)
	}
}
