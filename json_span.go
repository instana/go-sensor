package instana

import "github.com/opentracing/opentracing-go/ext"

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
	// HTTP server and client spans
	HTTPServerSpanType = RegisteredSpanType("g.http")
	HTTPClientSpanType = RegisteredSpanType("http")
	// RPC server and client spans
	RPCServerSpanType  = RegisteredSpanType("rpc-server")
	RPCClientSpanType  = RegisteredSpanType("rpc-client")
)

// RegisteredSpanType represents the span type supported by Instana
type RegisteredSpanType string

// ExtractData is a factory method to create the `data` section for a typed span
func (st RegisteredSpanType) ExtractData(span *spanS) typedSpanData {
	switch st {
	case HTTPServerSpanType, HTTPClientSpanType:
		return NewHTTPSpanData(span)
	case RPCServerSpanType, RPCClientSpanType:
		return NewRPCSpanData(span)
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

// HTTPSpanData represents the `data` section of an HTTP span sent within an OT span document
type HTTPSpanData struct {
	SpanData
	Tags HTTPSpanTags `json:"http"`
}

// NewHTTPSpanData initializes a new HTTP span data from tracer span
func NewHTTPSpanData(span *spanS) HTTPSpanData {
	data := HTTPSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     NewHTTPSpanTags(span),
	}

	return data
}

// HTTPSpanTags contains fields within the `data.http` section of an OT span document
type HTTPSpanTags struct {
	// Full request/response URL
	URL string `json:"url,omitempty"`
	// The HTTP status code returned with client/server response
	Status int `json:"status,omitempty"`
	// The HTTP method of the request
	Method string `json:"method,omitempty"`
	// Path is the path part of the request URL
	Path string `json:"path,omitempty"`
	// The name:port of the host to which the request had been sent
	Host string `json:"host,omitempty"`
	// The name of the protocol used for request ("http" or "https")
	Protocol string `json:"protocol,omitempty"`
	// The message describing an error occured during the request handling
	Error string `json:"error,omitempty"`
}

// NewHTTPSpanTags extracts HTTP-specific span tags from a tracer span
func NewHTTPSpanTags(span *spanS) HTTPSpanTags {
	var tags HTTPSpanTags
	for k, v := range span.Tags {
		switch k {
		case "http.url", string(ext.HTTPUrl):
			readStringTag(&tags.URL, v)
		case "http.status", "http.status_code":
			readIntTag(&tags.Status, v)
		case "http.method", string(ext.HTTPMethod):
			readStringTag(&tags.Method, v)
		case "http.path":
			readStringTag(&tags.Path, v)
		case "http.host":
			readStringTag(&tags.Host, v)
		case "http.protocol":
			readStringTag(&tags.Protocol, v)
		case "http.error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}

// RPCSpanData represents the `data` section of an RPC span sent within an OT span document
type RPCSpanData struct {
	SpanData
	Tags RPCSpanTags `json:"rpc"`
}

// NewRPCSpanData initializes a new RPC span data from tracer span
func NewRPCSpanData(span *spanS) RPCSpanData {
	data := RPCSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     NewRPCSpanTags(span),
	}

	return data
}

// RPCSpanTags contains fields within the `data.rpc` section of an OT span document
type RPCSpanTags struct {
	// The name of the remote host for an RPC call
	Host string `json:"host,omitempty"`
	// The port of the remote host for an RPC call
	Port string `json:"port,omitempty"`
	// The name of the remote method to invoke
	Call string `json:"call,omitempty"`
	// The type of an RPC call, e.g. either "unary" or "stream" for GRPC requests
	CallType string `json:"call_type,omitempty"`
	// The RPC flavor used for this call, e.g. "grpc" for GRPC requests
	Flavor string `json:"flavor,omitempty"`
	// The message describing an error occured during the request handling
	Error string `json:"error,omitempty"`
}

// NewRPCSpanTags extracts RPC-specific span tags from a tracer span
func NewRPCSpanTags(span *spanS) RPCSpanTags {
	var tags RPCSpanTags
	for k, v := range span.Tags {
		switch k {
		case "rpc.host":
			readStringTag(&tags.Host, v)
		case "rpc.port":
			readStringTag(&tags.Port, v)
		case "rpc.call":
			readStringTag(&tags.Call, v)
		case "rpc.call_type":
			readStringTag(&tags.CallType, v)
		case "rpc.flavor":
			readStringTag(&tags.Flavor, v)
		case "rpc.error":
			readStringTag(&tags.Error, v)
		}
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
