package instana

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/instana/go-sensor/w3ctrace"
	"github.com/opentracing/opentracing-go/ext"
)

type typedSpanData interface {
	Type() RegisteredSpanType
	Kind() SpanKind
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
	RPCServerSpanType = RegisteredSpanType("rpc-server")
	RPCClientSpanType = RegisteredSpanType("rpc-client")
	// Kafka consumer/producer span
	KafkaSpanType = RegisteredSpanType("kafka")
	// Google Cloud Storage client span
	GCPStorageSpanType = RegisteredSpanType("gcs")
	// Google Cloud PubSub client span
	GCPPubSubSpanType = RegisteredSpanType("gcps")
	// AWS Lambda entry span
	AWSLambdaEntrySpanType = RegisteredSpanType("aws.lambda.entry")
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
	case KafkaSpanType:
		return NewKafkaSpanData(span)
	case GCPStorageSpanType:
		return NewGCPStorageSpanData(span)
	case GCPPubSubSpanType:
		return NewGCPPubSubSpanData(span)
	case AWSLambdaEntrySpanType:
		return NewAWSLambdaSpanData(span)
	default:
		return NewSDKSpanData(span)
	}
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

// ForeignParent represents a related 3rd-party trace context, e.g. a W3C Trace Context
type ForeignParent struct {
	TraceID          string `json:"t"`
	ParentID         string `json:"p"`
	LatestTraceState string `json:"lts,omitempty"`
}

func newForeignParent(p interface{}) *ForeignParent {
	switch p := p.(type) {
	case w3ctrace.Context:
		return newW3CForeignParent(p)
	default:
		return nil
	}
}

func newW3CForeignParent(trCtx w3ctrace.Context) *ForeignParent {
	p, s := trCtx.Parent(), trCtx.State()

	var lastVendorData string
	if len(s) > 0 {
		lastVendorData = s[0]
	}

	return &ForeignParent{
		TraceID:          p.TraceID,
		ParentID:         p.ParentID,
		LatestTraceState: lastVendorData,
	}
}

// Span represents the OpenTracing span document to be sent to the agent
type Span struct {
	TraceID         int64
	TraceIDHi       int64
	ParentID        int64
	SpanID          int64
	Timestamp       uint64
	Duration        uint64
	Name            string
	From            *fromS
	Batch           *batchInfo
	Kind            int
	Ec              int
	Data            typedSpanData
	Synthetic       bool
	ForeignParent   *ForeignParent
	CorrelationType string
	CorrelationID   string
}

func newSpan(span *spanS) Span {
	data := RegisteredSpanType(span.Operation).ExtractData(span)
	sp := Span{
		TraceID:         span.context.TraceID,
		TraceIDHi:       span.context.TraceIDHi,
		ParentID:        span.context.ParentID,
		SpanID:          span.context.SpanID,
		Timestamp:       uint64(span.Start.UnixNano()) / uint64(time.Millisecond),
		Duration:        uint64(span.Duration) / uint64(time.Millisecond),
		Name:            string(data.Type()),
		Ec:              span.ErrorCount,
		ForeignParent:   newForeignParent(span.context.ForeignParent),
		CorrelationType: span.Correlation.Type,
		CorrelationID:   span.Correlation.ID,
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

	return sp
}

// MarshalJSON serializes span to JSON for sending it to Instana
func (sp Span) MarshalJSON() ([]byte, error) {
	var parentID string
	if sp.ParentID != 0 {
		parentID = FormatID(sp.ParentID)
	}

	return json.Marshal(struct {
		TraceID         string         `json:"t"`
		ParentID        string         `json:"p,omitempty"`
		SpanID          string         `json:"s"`
		Timestamp       uint64         `json:"ts"`
		Duration        uint64         `json:"d"`
		Name            string         `json:"n"`
		From            *fromS         `json:"f"`
		Batch           *batchInfo     `json:"b,omitempty"`
		Kind            int            `json:"k"`
		Ec              int            `json:"ec,omitempty"`
		Data            typedSpanData  `json:"data"`
		Synthetic       bool           `json:"sy,omitempty"`
		ForeignParent   *ForeignParent `json:"fp,omitempty"`
		CorrelationType string         `json:"crtp,omitempty"`
		CorrelationID   string         `json:"crid,omitempty"`
	}{
		FormatLongID(sp.TraceIDHi, sp.TraceID),
		parentID,
		FormatID(sp.SpanID),
		sp.Timestamp,
		sp.Duration,
		sp.Name,
		sp.From,
		sp.Batch,
		sp.Kind,
		sp.Ec,
		sp.Data,
		sp.Synthetic,
		sp.ForeignParent,
		sp.CorrelationType,
		sp.CorrelationID,
	})
}

type batchInfo struct {
	Size int `json:"s"`
}

// SpanData contains fields to be sent in the `data` section of an OT span document. These fields are
// common for all span types.
type SpanData struct {
	Service string `json:"service,omitempty"`
	st      RegisteredSpanType
	sk      interface{}
}

// NewSpanData initializes a new span data from tracer span
func NewSpanData(span *spanS, st RegisteredSpanType) SpanData {
	return SpanData{
		Service: span.Service,
		st:      st,
		sk:      span.Tags[string(ext.SpanKind)],
	}
}

// Type returns the registered span type suitable for use as the value of `n` field.
func (d SpanData) Type() RegisteredSpanType {
	return d.st
}

// Kind returns the kind of the span. It handles the github.com/opentracing/opentracing-go/ext.SpanKindEnum
// values as well as generic "entry" and "exit"
func (d SpanData) Kind() SpanKind {
	switch d.sk {
	case ext.SpanKindRPCServerEnum, string(ext.SpanKindRPCServerEnum),
		ext.SpanKindConsumerEnum, string(ext.SpanKindConsumerEnum),
		"entry":
		return EntrySpanKind
	case ext.SpanKindRPCClientEnum, string(ext.SpanKindRPCClientEnum),
		ext.SpanKindProducerEnum, string(ext.SpanKindProducerEnum),
		"exit":
		return ExitSpanKind
	default:
		return IntermediateSpanKind
	}
}

// SDKSpanData represents the `data` section of an SDK span sent within an OT span document
type SDKSpanData struct {
	SpanData
	Tags SDKSpanTags `json:"sdk"`
}

// NewSDKSpanData initializes a new SDK span data from a tracer span
func NewSDKSpanData(span *spanS) SDKSpanData {
	d := NewSpanData(span, SDKSpanType)
	return SDKSpanData{
		SpanData: d,
		Tags:     NewSDKSpanTags(span, d.Kind().String()),
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
func NewSDKSpanTags(span *spanS, spanType string) SDKSpanTags {
	tags := SDKSpanTags{
		Name:   span.Operation,
		Type:   spanType,
		Custom: map[string]interface{}{},
	}

	if len(span.Tags) != 0 {
		tags.Custom["tags"] = span.Tags
	}

	if logs := collectTracerSpanLogs(span); len(logs) > 0 {
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
	// Params are the request query string parameters
	Params string `json:"params,omitempty"`
	// Headers are the captured request/response headers
	Headers map[string]string `json:"header,omitempty"`
	// PathTemplate is the raw template string used to route the request
	PathTemplate string `json:"path_tpl,omitempty"`
	// The name:port of the host to which the request had been sent
	Host string `json:"host,omitempty"`
	// The name of the protocol used for request ("http" or "https")
	Protocol string `json:"protocol,omitempty"`
	// The message describing an error occurred during the request handling
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
		case "http.params":
			readStringTag(&tags.Params, v)
		case "http.header":
			if m, ok := v.(map[string]string); ok {
				tags.Headers = m
			}
		case "http.path_tpl":
			readStringTag(&tags.PathTemplate, v)
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
	// The message describing an error occurred during the request handling
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

// KafkaSpanData represents the `data` section of an Kafka span sent within an OT span document
type KafkaSpanData struct {
	SpanData
	Tags KafkaSpanTags `json:"kafka"`
}

// NewKafkaSpanData initializes a new Kafka span data from tracer span
func NewKafkaSpanData(span *spanS) KafkaSpanData {
	data := KafkaSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     NewKafkaSpanTags(span),
	}

	return data
}

// KafkaSpanTags contains fields within the `data.kafka` section of an OT span document
type KafkaSpanTags struct {
	// Kafka topic
	Service string `json:"service"`
	// The access mode:, either "send" for publisher or "consume" for consumer
	Access string `json:"access"`
}

// NewKafkaSpanTags extracts Kafka-specific span tags from a tracer span
func NewKafkaSpanTags(span *spanS) KafkaSpanTags {
	var tags KafkaSpanTags
	for k, v := range span.Tags {
		switch k {
		case "kafka.service":
			readStringTag(&tags.Service, v)
		case "kafka.access":
			readStringTag(&tags.Access, v)
		}
	}

	return tags
}

// GCPStorageSpanData represents the `data` section of a Google Cloud Storage span sent within an OT span document
type GCPStorageSpanData struct {
	SpanData
	Tags GCPStorageSpanTags `json:"gcs"`
}

// NewGCPStorageSpanData initializes a new Google Cloud Storage span data from tracer span
func NewGCPStorageSpanData(span *spanS) GCPStorageSpanData {
	data := GCPStorageSpanData{
		SpanData: NewSpanData(span, GCPStorageSpanType),
		Tags:     NewGCPStorageSpanTags(span),
	}

	return data
}

// Kind returns the span kind for a Google Cloud Storage span
func (d GCPStorageSpanData) Kind() SpanKind {
	return ExitSpanKind
}

// GCPStorageSpanTags contains fields within the `data.gcs` section of an OT span document
type GCPStorageSpanTags struct {
	Operation          string `json:"op,omitempty"`
	Bucket             string `json:"bucket,omitempty"`
	Object             string `json:"object,omitempty"`
	Entity             string `json:"entity,omitempty"`
	Range              string `json:"range,omitempty"`
	SourceBucket       string `json:"sourceBucket,omitempty"`
	SourceObject       string `json:"sourceObject,omitempty"`
	DestinationBucket  string `json:"destinationBucket,omitempty"`
	DestinationObject  string `json:"destinationObject,omitempty"`
	NumberOfOperations string `json:"numberOfOperations,omitempty"`
	ProjectID          string `json:"projectId,omitempty"`
	AccessID           string `json:"accessId,omitempty"`
}

// NewGCPStorageSpanTags extracts Google Cloud Storage span tags from a tracer span
func NewGCPStorageSpanTags(span *spanS) GCPStorageSpanTags {
	var tags GCPStorageSpanTags
	for k, v := range span.Tags {
		switch k {
		case "gcs.op":
			readStringTag(&tags.Operation, v)
		case "gcs.bucket":
			readStringTag(&tags.Bucket, v)
		case "gcs.object":
			readStringTag(&tags.Object, v)
		case "gcs.entity":
			readStringTag(&tags.Entity, v)
		case "gcs.range":
			readStringTag(&tags.Range, v)
		case "gcs.sourceBucket":
			readStringTag(&tags.SourceBucket, v)
		case "gcs.sourceObject":
			readStringTag(&tags.SourceObject, v)
		case "gcs.destinationBucket":
			readStringTag(&tags.DestinationBucket, v)
		case "gcs.destinationObject":
			readStringTag(&tags.DestinationObject, v)
		case "gcs.numberOfOperations":
			readStringTag(&tags.NumberOfOperations, v)
		case "gcs.projectId":
			readStringTag(&tags.ProjectID, v)
		case "gcs.accessId":
			readStringTag(&tags.AccessID, v)
		}
	}

	return tags
}

// GCPPubSubSpanData represents the `data` section of a Google Cloud Pub/Sub span sent within an OT span document
type GCPPubSubSpanData struct {
	SpanData
	Tags GCPPubSubSpanTags `json:"gcps"`
}

// NewGCPPubSubSpanData initializes a new Google Cloud Pub/Span span data from tracer span
func NewGCPPubSubSpanData(span *spanS) GCPPubSubSpanData {
	data := GCPPubSubSpanData{
		SpanData: NewSpanData(span, GCPPubSubSpanType),
		Tags:     NewGCPPubSubSpanTags(span),
	}

	return data
}

// Kind returns the span kind for a Google Cloud Pub/Sub span
func (d GCPPubSubSpanData) Kind() SpanKind {
	switch strings.ToLower(d.Tags.Operation) {
	case "consume":
		return EntrySpanKind
	default:
		return ExitSpanKind
	}
}

// GCPPubSubSpanTags contains fields within the `data.gcps` section of an OT span document
type GCPPubSubSpanTags struct {
	ProjectID    string `json:"projid"`
	Operation    string `json:"op"`
	Topic        string `json:"top,omitempty"`
	Subscription string `json:"sub,omitempty"`
	MessageID    string `json:"msgid,omitempty"`
}

// NewGCPPubSubSpanTags extracts Google Cloud Pub/Sub span tags from a tracer span
func NewGCPPubSubSpanTags(span *spanS) GCPPubSubSpanTags {
	var tags GCPPubSubSpanTags
	for k, v := range span.Tags {
		switch k {
		case "gcps.projid":
			readStringTag(&tags.ProjectID, v)
		case "gcps.op":
			readStringTag(&tags.Operation, v)
		case "gcps.top":
			readStringTag(&tags.Topic, v)
		case "gcps.sub":
			readStringTag(&tags.Subscription, v)
		case "gcps.msgid":
			readStringTag(&tags.MessageID, v)
		}
	}

	return tags
}

// AWSLambdaCloudWatchSpanTags contains fields within the `data.lambda.cw` section of an OT span document
type AWSLambdaCloudWatchSpanTags struct {
	Events *AWSLambdaCloudWatchEventTags `json:"events,omitempty"`
	Logs   *AWSLambdaCloudWatchLogsTags  `json:"logs,omitempty"`
}

// NewAWSLambdaCloudWatchSpanTags extracts CloudWatch tags for an AWS Lambda entry span
func NewAWSLambdaCloudWatchSpanTags(span *spanS) AWSLambdaCloudWatchSpanTags {
	var tags AWSLambdaCloudWatchSpanTags

	if events := NewAWSLambdaCloudWatchEventTags(span); !events.IsZero() {
		tags.Events = &events
	}

	if logs := NewAWSLambdaCloudWatchLogsTags(span); !logs.IsZero() {
		tags.Logs = &logs
	}

	return tags
}

// IsZero returns true if an AWSLambdaCloudWatchSpanTags struct was populated with event data
func (tags AWSLambdaCloudWatchSpanTags) IsZero() bool {
	return (tags.Events == nil || tags.Events.IsZero()) && (tags.Logs == nil || tags.Logs.IsZero())
}

// AWSLambdaCloudWatchEventTags contains fields within the `data.lambda.cw.events` section of an OT span document
type AWSLambdaCloudWatchEventTags struct {
	// ID is the ID of the event
	ID string `json:"id"`
	// Resources contains the event resources
	Resources []string `json:"resources"`
	// More is set to true if the event resources list was truncated
	More bool `json:"more,omitempty"`
}

// NewAWSLambdaCloudWatchEventTags extracts CloudWatch event tags for an AWS Lambda entry span. It truncates
// the resources list to the first 3 items, populating the `data.lambda.cw.events.more` tag and limits each
// resource string to the first 200 characters to reduce the payload.
func NewAWSLambdaCloudWatchEventTags(span *spanS) AWSLambdaCloudWatchEventTags {
	var tags AWSLambdaCloudWatchEventTags

	if v, ok := span.Tags["cloudwatch.events.id"]; ok {
		readStringTag(&tags.ID, v)
	}

	if v, ok := span.Tags["cloudwatch.events.resources"]; ok {
		switch v := v.(type) {
		case []string:
			if len(v) > 3 {
				v = v[:3]
				tags.More = true
			}

			tags.Resources = v
		case string:
			tags.Resources = []string{v}
		case []byte:
			tags.Resources = []string{string(v)}
		}
	}

	// truncate resources
	if len(tags.Resources) > 3 {
		tags.Resources, tags.More = tags.Resources[:3], true
	}

	for i := range tags.Resources {
		if len(tags.Resources[i]) > 200 {
			tags.Resources[i] = tags.Resources[i][:200]
		}
	}

	return tags
}

// IsZero returns true if an AWSCloudWatchEventTags struct was populated with event data
func (tags AWSLambdaCloudWatchEventTags) IsZero() bool {
	return tags.ID == ""
}

// AWSLambdaCloudWatchLogsTags contains fields within the `data.lambda.cw.logs` section of an OT span document
type AWSLambdaCloudWatchLogsTags struct {
	Group         string   `json:"group"`
	Stream        string   `json:"stream"`
	Events        []string `json:"events"`
	More          bool     `json:"more,omitempty"`
	DecodingError string   `json:"decodingError,omitempty"`
}

// NewAWSLambdaCloudWatchLogsTags extracts CloudWatch Logs tags for an AWS Lambda entry span. It truncates
// the log events list to the first 3 items, populating the `data.lambda.cw.logs.more` tag and limits each
// log string to the first 200 characters to reduce the payload.
func NewAWSLambdaCloudWatchLogsTags(span *spanS) AWSLambdaCloudWatchLogsTags {
	var tags AWSLambdaCloudWatchLogsTags

	if v, ok := span.Tags["cloudwatch.logs.group"]; ok {
		readStringTag(&tags.Group, v)
	}

	if v, ok := span.Tags["cloudwatch.logs.stream"]; ok {
		readStringTag(&tags.Stream, v)
	}

	if v, ok := span.Tags["cloudwatch.logs.decodingError"]; ok {
		switch v := v.(type) {
		case error:
			tags.DecodingError = v.Error()
		case string:
			tags.DecodingError = v
		}
	}

	if v, ok := span.Tags["cloudwatch.logs.events"]; ok {
		switch v := v.(type) {
		case []string:
			if len(v) > 3 {
				v = v[:3]
				tags.More = true
			}

			tags.Events = v
		case string:
			tags.Events = []string{v}
		case []byte:
			tags.Events = []string{string(v)}
		}
	}

	// truncate events
	if len(tags.Events) > 3 {
		tags.Events, tags.More = tags.Events[:3], true
	}

	for i := range tags.Events {
		if len(tags.Events[i]) > 200 {
			tags.Events[i] = tags.Events[i][:200]
		}
	}

	return tags
}

// IsZero returns true if an AWSLambdaCloudWatchLogsTags struct was populated with logs data
func (tags AWSLambdaCloudWatchLogsTags) IsZero() bool {
	return tags.Group == "" && tags.Stream == "" && tags.DecodingError == ""
}

// AWSS3EventTags represens metadata for an S3 event
type AWSS3EventTags struct {
	Name   string `json:"event"`
	Bucket string `json:"bucket"`
	Object string `json:"object,omitempty"`
}

// AWSLambdaS3SpanTags contains fields within the `data.lambda.s3` section of an OT span document
type AWSLambdaS3SpanTags struct {
	Events []AWSS3EventTags `json:"events,omitempty"`
}

// NewAWSLambdaS3SpanTags extracts S3 Event tags for an AWS Lambda entry span. It truncates
// the events list to the first 3 items and limits each object names to the first 200 characters to reduce the payload.
func NewAWSLambdaS3SpanTags(span *spanS) AWSLambdaS3SpanTags {
	var tags AWSLambdaS3SpanTags

	if events, ok := span.Tags["s3.events"]; ok {
		events, ok := events.([]AWSS3EventTags)
		if ok {
			tags.Events = events
		}
	}

	if len(tags.Events) > 3 {
		tags.Events = tags.Events[:3]
	}

	for i := range tags.Events {
		if len(tags.Events[i].Object) > 200 {
			tags.Events[i].Object = tags.Events[i].Object[:200]
		}
	}

	return tags
}

// IsZero returns true if an AWSLambdaS3SpanTags struct was populated with events data
func (tags AWSLambdaS3SpanTags) IsZero() bool {
	return len(tags.Events) == 0
}

// AWSSQSMessageTags represents span tags for an SQS message delivery
type AWSSQSMessageTags struct {
	Queue string `json:"queue"`
}

// AWSLambdaSQSSpanTags contains fields within the `data.lambda.sqs` section of an OT span document
type AWSLambdaSQSSpanTags struct {
	// Messages are message tags for an SQS event
	Messages []AWSSQSMessageTags `json:"messages"`
}

// NewAWSLambdaSQSSpanTags extracts SQS event tags for an AWS Lambda entry span. It truncates
// the events list to the first 3 items to reduce the payload.
func NewAWSLambdaSQSSpanTags(span *spanS) AWSLambdaSQSSpanTags {
	var tags AWSLambdaSQSSpanTags

	if msgs, ok := span.Tags["sqs.messages"]; ok {
		msgs, ok := msgs.([]AWSSQSMessageTags)
		if ok {
			tags.Messages = msgs
		}
	}

	if len(tags.Messages) > 3 {
		tags.Messages = tags.Messages[:3]
	}

	return tags
}

// IsZero returns true if an AWSLambdaSQSSpanTags struct was populated with messages data
func (tags AWSLambdaSQSSpanTags) IsZero() bool {
	return len(tags.Messages) == 0
}

// AWSLambdaSpanTags contains fields within the `data.lambda` section of an OT span document
type AWSLambdaSpanTags struct {
	// ARN is the ARN of invoked AWS Lambda function with the version attached
	ARN string `json:"arn"`
	// Runtime is an Instana constant for this AWS lambda runtime (always "go")
	Runtime string `json:"runtime"`
	// Name is the name of invoked function
	Name string `json:"functionName,omitempty"`
	// Version is either the numeric version or $LATEST
	Version string `json:"functionVersion,omitempty"`
	// Trigger is the trigger event type (if any)
	Trigger string `json:"trigger,omitempty"`
	// CloudWatch holds the details of a CloudWatch event associated with this lambda
	CloudWatch *AWSLambdaCloudWatchSpanTags `json:"cw,omitempty"`
	// S3 holds the details of a S3 events associated with this lambda
	S3 *AWSLambdaS3SpanTags
	// SQS holds the details of a SQS events associated with this lambda
	SQS *AWSLambdaSQSSpanTags
}

// NewAWSLambdaSpanTags extracts AWS Lambda entry span tags from a tracer span
func NewAWSLambdaSpanTags(span *spanS) AWSLambdaSpanTags {
	tags := AWSLambdaSpanTags{Runtime: "go"}

	if v, ok := span.Tags["lambda.arn"]; ok {
		readStringTag(&tags.ARN, v)
	}

	if v, ok := span.Tags["lambda.name"]; ok {
		readStringTag(&tags.Name, v)
	}

	if v, ok := span.Tags["lambda.version"]; ok {
		readStringTag(&tags.Version, v)
	}

	if v, ok := span.Tags["lambda.trigger"]; ok {
		readStringTag(&tags.Trigger, v)
	}

	if cw := NewAWSLambdaCloudWatchSpanTags(span); !cw.IsZero() {
		tags.CloudWatch = &cw
	}

	if st := NewAWSLambdaS3SpanTags(span); !st.IsZero() {
		tags.S3 = &st
	}

	if sqs := NewAWSLambdaSQSSpanTags(span); !sqs.IsZero() {
		tags.SQS = &sqs
	}

	return tags
}

// AWSLambdaSpanData is the base span data type for AWS Lambda entry spans
type AWSLambdaSpanData struct {
	Snapshot AWSLambdaSpanTags `json:"lambda"`
	HTTP     *HTTPSpanTags     `json:"http,omitempty"`
}

// NewAWSLambdaSpanData initializes a new AWSLambdaSpanData from span
func NewAWSLambdaSpanData(span *spanS) AWSLambdaSpanData {
	d := AWSLambdaSpanData{
		Snapshot: NewAWSLambdaSpanTags(span),
	}

	switch span.Tags["lambda.trigger"] {
	case "aws:api.gateway", "aws:application.load.balancer":
		tags := NewHTTPSpanTags(span)
		d.HTTP = &tags
	}

	return d
}

// Type returns the span type for an AWS Lambda span
func (d AWSLambdaSpanData) Type() RegisteredSpanType {
	return AWSLambdaEntrySpanType
}

// Kind returns the span kind for an AWS Lambda span
func (d AWSLambdaSpanData) Kind() SpanKind {
	return EntrySpanKind
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

func collectTracerSpanLogs(span *spanS) map[uint64]map[string]interface{} {
	logs := make(map[uint64]map[string]interface{})
	for _, l := range span.Logs {
		if _, ok := logs[uint64(l.Timestamp.UnixNano())/uint64(time.Millisecond)]; !ok {
			logs[uint64(l.Timestamp.UnixNano())/uint64(time.Millisecond)] = make(map[string]interface{})
		}

		for _, f := range l.Fields {
			logs[uint64(l.Timestamp.UnixNano())/uint64(time.Millisecond)][f.Key()] = f.Value()
		}
	}

	return logs
}
