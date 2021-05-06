// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instana

import (
	"strings"

	"github.com/opentracing/opentracing-go/ext"
)

// Registered types supported by Instana. The span type is determined based on
// the operation name passed to the `StartSpan()` call of a tracer.
//
// It is NOT RECOMMENDED to use operation names that match any of these constants in your
// custom instrumentation code unless you explicitly wish to send data as a registered span.
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
	// AWS S3 client span
	AWSS3SpanType = RegisteredSpanType("s3")
	// AWS SQS client span
	AWSSQSSpanType = RegisteredSpanType("sqs")
	// AWS SNS client span
	AWSSNSSpanType = RegisteredSpanType("sns")
	// AWS DynamoDB client span
	AWSDynamoDBSpanType = RegisteredSpanType("dynamodb")
	// AWS Lambda invoke span
	AWSInvokeSpanType = RegisteredSpanType("invoke")
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
	case AWSS3SpanType:
		return NewAWSS3SpanData(span)
	case AWSSQSSpanType:
		return NewAWSSQSSpanData(span)
	case AWSSNSSpanType:
		return NewAWSSNSSpanData(span)
	case AWSDynamoDBSpanType:
		return NewAWSDynamoDBSpanData(span)
	case AWSInvokeSpanType:
		return newAWSInvokeSpanData(span)
	default:
		return NewSDKSpanData(span)
	}
}

// TagsNames returns a set of tag names know to the registered span type
func (st RegisteredSpanType) TagsNames() map[string]struct{} {
	var yes struct{}

	switch st {
	case HTTPServerSpanType, HTTPClientSpanType:
		return map[string]struct{}{
			"http.url": yes, string(ext.HTTPUrl): yes,
			"http.status": yes, "http.status_code": yes,
			"http.method": yes, string(ext.HTTPMethod): yes,
			"http.path":     yes,
			"http.params":   yes,
			"http.header":   yes,
			"http.path_tpl": yes,
			"http.host":     yes,
			"http.protocol": yes,
			"http.error":    yes,
		}
	case RPCServerSpanType, RPCClientSpanType:
		return map[string]struct{}{
			"rpc.host":      yes,
			"rpc.port":      yes,
			"rpc.call":      yes,
			"rpc.call_type": yes,
			"rpc.flavor":    yes,
			"rpc.error":     yes,
		}
	case KafkaSpanType:
		return map[string]struct{}{
			"kafka.service": yes,
			"kafka.access":  yes,
		}
	case GCPStorageSpanType:
		return map[string]struct{}{
			"gcs.op":                 yes,
			"gcs.bucket":             yes,
			"gcs.object":             yes,
			"gcs.entity":             yes,
			"gcs.range":              yes,
			"gcs.sourceBucket":       yes,
			"gcs.sourceObject":       yes,
			"gcs.destinationBucket":  yes,
			"gcs.destinationObject":  yes,
			"gcs.numberOfOperations": yes,
			"gcs.projectId":          yes,
			"gcs.accessId":           yes,
		}
	case GCPPubSubSpanType:
		return map[string]struct{}{
			"gcps.projid": yes,
			"gcps.op":     yes,
			"gcps.top":    yes,
			"gcps.sub":    yes,
			"gcps.msgid":  yes,
		}
	case AWSLambdaEntrySpanType:
		return map[string]struct{}{
			"lambda.arn":                    yes,
			"lambda.name":                   yes,
			"lambda.version":                yes,
			"lambda.trigger":                yes,
			"cloudwatch.events.id":          yes,
			"cloudwatch.events.resources":   yes,
			"cloudwatch.logs.group":         yes,
			"cloudwatch.logs.stream":        yes,
			"cloudwatch.logs.decodingError": yes,
			"cloudwatch.logs.events":        yes,
			"s3.events":                     yes,
			"sqs.messages":                  yes,
		}
	case AWSS3SpanType:
		return map[string]struct{}{
			"s3.region": yes,
			"s3.op":     yes,
			"s3.bucket": yes,
			"s3.key":    yes,
			"s3.error":  yes,
		}
	case AWSSQSSpanType:
		return map[string]struct{}{
			"sqs.sort":  yes,
			"sqs.queue": yes,
			"sqs.type":  yes,
			"sqs.group": yes,
			"sqs.size":  yes,
			"sqs.error": yes,
		}
	case AWSSNSSpanType:
		return map[string]struct{}{
			"sns.topic":   yes,
			"sns.target":  yes,
			"sns.phone":   yes,
			"sns.subject": yes,
			"sns.error":   yes,
		}
	case AWSDynamoDBSpanType:
		return map[string]struct{}{
			"dynamodb.table": yes,
			"dynamodb.op":    yes,
			"dynamodb.error": yes,
		}
	case AWSInvokeSpanType:
		return map[string]struct{}{
			"invoke.function": yes,
			"invoke.type":     yes,
			"invoke.error":    yes,
		}
	default:
		return nil
	}
}

// HTTPSpanData represents the `data` section of an HTTP span sent within an OT span document
type HTTPSpanData struct {
	SpanData
	Tags HTTPSpanTags `json:"http"`

	clientSpan bool
}

// NewHTTPSpanData initializes a new HTTP span data from tracer span
func NewHTTPSpanData(span *spanS) HTTPSpanData {
	data := HTTPSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     NewHTTPSpanTags(span),
	}

	kindTag := span.Tags[string(ext.SpanKind)]
	data.clientSpan = kindTag == ext.SpanKindRPCClientEnum || kindTag == string(ext.SpanKindRPCClientEnum)

	return data
}

// Kind returns instana.EntrySpanKind for server spans and instana.ExitSpanKind otherwise
func (d HTTPSpanData) Kind() SpanKind {
	if d.clientSpan {
		return ExitSpanKind
	}

	return EntrySpanKind
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

	clientSpan bool
}

// NewRPCSpanData initializes a new RPC span data from tracer span
func NewRPCSpanData(span *spanS) RPCSpanData {
	data := RPCSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     NewRPCSpanTags(span),
	}

	kindTag := span.Tags[string(ext.SpanKind)]
	data.clientSpan = kindTag == ext.SpanKindRPCClientEnum || kindTag == (ext.SpanKindRPCClientEnum)

	return data
}

// Kind returns instana.EntrySpanKind for server spans and instana.ExitSpanKind otherwise
func (d RPCSpanData) Kind() SpanKind {
	if d.clientSpan {
		return ExitSpanKind
	}

	return EntrySpanKind
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

	producerSpan bool
}

// NewKafkaSpanData initializes a new Kafka span data from tracer span
func NewKafkaSpanData(span *spanS) KafkaSpanData {
	data := KafkaSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     NewKafkaSpanTags(span),
	}

	kindTag := span.Tags[string(ext.SpanKind)]
	data.producerSpan = kindTag == ext.SpanKindProducerEnum || kindTag == string(ext.SpanKindProducerEnum)

	return data
}

func (d KafkaSpanData) Kind() SpanKind {
	if d.producerSpan {
		return ExitSpanKind
	}

	return EntrySpanKind
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

// AWSS3SpanData represents the `data` section of a AWS S3 span sent within an OT span document
type AWSS3SpanData struct {
	SpanData
	Tags AWSS3SpanTags `json:"s3"`
}

// NewAWSS3SpanData initializes a new AWS S3 span data from tracer span
func NewAWSS3SpanData(span *spanS) AWSS3SpanData {
	data := AWSS3SpanData{
		SpanData: NewSpanData(span, AWSS3SpanType),
		Tags:     NewAWSS3SpanTags(span),
	}

	return data
}

// Kind returns the span kind for a AWS S3 span
func (d AWSS3SpanData) Kind() SpanKind {
	return ExitSpanKind
}

// AWSS3SpanTags contains fields within the `data.s3` section of an OT span document
type AWSS3SpanTags struct {
	// Region is the AWS region used to access S3
	Region string `json:"region,omitempty"`
	// Operation is the operation name, as defined by AWS S3 API
	Operation string `json:"op,omitempty"`
	// Bucket is the bucket name
	Bucket string `json:"bucket,omitempty"`
	// Key is the object key
	Key string `json:"key,omitempty"`
	// Error is an optional error returned by AWS API
	Error string `json:"error,omitempty"`
}

// NewAWSS3SpanTags extracts AWS S3 span tags from a tracer span
func NewAWSS3SpanTags(span *spanS) AWSS3SpanTags {
	var tags AWSS3SpanTags
	for k, v := range span.Tags {
		switch k {
		case "s3.region":
			readStringTag(&tags.Region, v)
		case "s3.op":
			readStringTag(&tags.Operation, v)
		case "s3.bucket":
			readStringTag(&tags.Bucket, v)
		case "s3.key":
			readStringTag(&tags.Key, v)
		case "s3.error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}

// AWSSQSSpanData represents the `data` section of a AWS SQS span sent within an OT span document
type AWSSQSSpanData struct {
	SpanData
	Tags AWSSQSSpanTags `json:"sqs"`
}

// NewAWSSQSSpanData initializes a new AWS SQS span data from tracer span
func NewAWSSQSSpanData(span *spanS) AWSSQSSpanData {
	data := AWSSQSSpanData{
		SpanData: NewSpanData(span, AWSSQSSpanType),
		Tags:     NewAWSSQSSpanTags(span),
	}

	return data
}

// Kind returns the span kind for a AWS SQS span
func (d AWSSQSSpanData) Kind() SpanKind {
	switch d.Tags.Sort {
	case "entry":
		return EntrySpanKind
	case "exit":
		return ExitSpanKind
	default:
		return IntermediateSpanKind
	}
}

// AWSSQSSpanTags contains fields within the `data.sqs` section of an OT span document
type AWSSQSSpanTags struct {
	// Sort is the direction of the call, wither "entry" or "exit"
	Sort string `json:"sort,omitempty"`
	// Queue is the queue name
	Queue string `json:"queue,omitempty"`
	// Type is the operation name
	Type string `json:"type,omitempty"`
	// MessageGroupID is the message group ID specified while sending messages
	MessageGroupID string `json:"group,omitempty"`
	// Size is the optional batch size
	Size int `json:"size,omitempty"`
	// Error is an optional error returned by AWS API
	Error string `json:"error,omitempty"`
}

// NewAWSSQSSpanTags extracts AWS SQS span tags from a tracer span
func NewAWSSQSSpanTags(span *spanS) AWSSQSSpanTags {
	var tags AWSSQSSpanTags
	for k, v := range span.Tags {
		switch k {
		case "sqs.sort":
			readStringTag(&tags.Sort, v)
		case "sqs.queue":
			readStringTag(&tags.Queue, v)
		case "sqs.type":
			readStringTag(&tags.Type, v)
		case "sqs.group":
			readStringTag(&tags.MessageGroupID, v)
		case "sqs.size":
			readIntTag(&tags.Size, v)
		case "sqs.error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}

// AWSSNSSpanData represents the `data` section of a AWS SNS span sent within an OT span document
type AWSSNSSpanData struct {
	SpanData
	Tags AWSSNSSpanTags `json:"sns"`
}

// NewAWSSNSSpanData initializes a new AWS SNS span data from tracer span
func NewAWSSNSSpanData(span *spanS) AWSSNSSpanData {
	data := AWSSNSSpanData{
		SpanData: NewSpanData(span, AWSSNSSpanType),
		Tags:     NewAWSSNSSpanTags(span),
	}

	return data
}

// Kind returns the span kind for a AWS SNS span
func (d AWSSNSSpanData) Kind() SpanKind {
	return ExitSpanKind
}

// AWSSNSSpanTags contains fields within the `data.sns` section of an OT span document
type AWSSNSSpanTags struct {
	// TopicARN is the topic ARN of an SNS message
	TopicARN string `json:"topic,omitempty"`
	// TargetARN is the target ARN of an SNS message
	TargetARN string `json:"target,omitempty"`
	// Phone is the phone no. of an SNS message
	Phone string `json:"phone,omitempty"`
	// Subject is the subject of an SNS message
	Subject string `json:"subject,omitempty"`
	// Error is an optional error returned by AWS API
	Error string `json:"error,omitempty"`
}

// NewAWSSNSSpanTags extracts AWS SNS span tags from a tracer span
func NewAWSSNSSpanTags(span *spanS) AWSSNSSpanTags {
	var tags AWSSNSSpanTags
	for k, v := range span.Tags {
		switch k {
		case "sns.topic":
			readStringTag(&tags.TopicARN, v)
		case "sns.target":
			readStringTag(&tags.TargetARN, v)
		case "sns.phone":
			readStringTag(&tags.Phone, v)
		case "sns.subject":
			readStringTag(&tags.Subject, v)
		case "sns.error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}

// AWSDynamoDBSpanData represents the `data` section of a AWS DynamoDB span sent within an OT span document
type AWSDynamoDBSpanData struct {
	SpanData
	Tags AWSDynamoDBSpanTags `json:"sns"`
}

// NewAWSDynamoDBSpanData initializes a new AWS DynamoDB span data from tracer span
func NewAWSDynamoDBSpanData(span *spanS) AWSDynamoDBSpanData {
	data := AWSDynamoDBSpanData{
		SpanData: NewSpanData(span, AWSDynamoDBSpanType),
		Tags:     NewAWSDynamoDBSpanTags(span),
	}

	return data
}

// Kind returns the span kind for a AWS DynamoDB span
func (d AWSDynamoDBSpanData) Kind() SpanKind {
	return ExitSpanKind
}

// AWSDynamoDBSpanTags contains fields within the `data.sns` section of an OT span document
type AWSDynamoDBSpanTags struct {
	// Table is the name of DynamoDB table
	Table string `json:"table,omitempty"`
	// Operation is the operation name
	Operation string `json:"op,omitempty"`
	// Error is an optional name returned by AWS API
	Error string `json:"error,omitempty"`
}

// NewAWSDynamoDBSpanTags extracts AWS DynamoDB span tags from a tracer span
func NewAWSDynamoDBSpanTags(span *spanS) AWSDynamoDBSpanTags {
	var tags AWSDynamoDBSpanTags
	for k, v := range span.Tags {
		switch k {
		case "dynamodb.table":
			readStringTag(&tags.Table, v)
		case "dynamodb.op":
			readStringTag(&tags.Operation, v)
		case "dynamodb.error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}

type AWSInvokeSpanTags struct {
	// FunctionName is a name of the function which is invoked
	FunctionName string `json:"function,omitempty"`
	// InvocationType if equal to `Event`, means it is an async invocation
	InvocationType string `json:"type,omitempty"`
	// Error is an optional error returned by AWS API
	Error string `json:"error,omitempty"`
}

func newAWSDInvokeSpanTags(span *spanS) AWSInvokeSpanTags {
	var tags AWSInvokeSpanTags
	for k, v := range span.Tags {
		switch k {
		case "invoke.function":
			readStringTag(&tags.FunctionName, v)
		case "invoke.type":
			readStringTag(&tags.InvocationType, v)
		case "invoke.error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}

type AWSInvokeSpanData struct {
	SpanData
	Tags AWSInvokeSpanTags `json:"invoke"`
}

// Kind returns the span kind for a AWS Invoke span
func (d AWSInvokeSpanData) Kind() SpanKind {
	return ExitSpanKind
}

// Type returns the span type for an AWS Invoke span
func (d AWSInvokeSpanData) Type() RegisteredSpanType {
	return AWSInvokeSpanType
}

// newAWSInvokeSpanData initializes a new AWS Invoke span data from tracer span
func newAWSInvokeSpanData(span *spanS) AWSInvokeSpanData {
	data := AWSInvokeSpanData{
		SpanData: NewSpanData(span, AWSInvokeSpanType),
		Tags:     newAWSDInvokeSpanTags(span),
	}

	return data
}
