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
	AWSLambdaInvokeSpanType = RegisteredSpanType("aws.lambda.invoke")
	// Logging span
	LogSpanType = RegisteredSpanType("log.go")
	// MongoDB client span
	MongoDBSpanType = RegisteredSpanType("mongo")
	// PostgreSQL client span
	PostgreSQLSpanType = RegisteredSpanType("postgres")
	// Redis client span
	RedisSpanType = RegisteredSpanType("redis")
	// RabbitMQ client span
	RabbitMQSpanType = RegisteredSpanType("rabbitmq")
	// Azure function span
	AzureFunctionType = RegisteredSpanType("azf")
	// GraphQL server span
	GraphQLServerType = RegisteredSpanType("graphql.server")
	// GraphQL client span
	GraphQLClientType = RegisteredSpanType("graphql.client")
)

// RegisteredSpanType represents the span type supported by Instana
type RegisteredSpanType string

// extractData is a factory method to create the `data` section for a typed span
func (st RegisteredSpanType) extractData(span *spanS) typedSpanData {
	switch st {
	case HTTPServerSpanType, HTTPClientSpanType:
		return newHTTPSpanData(span)
	case RPCServerSpanType, RPCClientSpanType:
		return newRPCSpanData(span)
	case KafkaSpanType:
		return newKafkaSpanData(span)
	case GCPStorageSpanType:
		return newGCPStorageSpanData(span)
	case GCPPubSubSpanType:
		return newGCPPubSubSpanData(span)
	case AWSLambdaEntrySpanType:
		return newAWSLambdaSpanData(span)
	case AWSS3SpanType:
		return newAWSS3SpanData(span)
	case AWSSQSSpanType:
		return newAWSSQSSpanData(span)
	case AWSSNSSpanType:
		return newAWSSNSSpanData(span)
	case AWSDynamoDBSpanType:
		return newAWSDynamoDBSpanData(span)
	case AWSLambdaInvokeSpanType:
		return newAWSLambdaInvokeSpanData(span)
	case LogSpanType:
		return newLogSpanData(span)
	case MongoDBSpanType:
		return newMongoDBSpanData(span)
	case PostgreSQLSpanType:
		return newPostgreSQLSpanData(span)
	case RedisSpanType:
		return newRedisSpanData(span)
	case RabbitMQSpanType:
		return newRabbitMQSpanData(span)
	case AzureFunctionType:
		return newAZFSpanData(span)
	case GraphQLServerType, GraphQLClientType:
		return newGraphQLSpanData(span)
	default:
		return newSDKSpanData(span)
	}
}

// TagsNames returns a set of tag names known to the registered span type
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
			"http.route_id": yes,
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
			"lambda.coldStart":              yes,
			"lambda.msleft":                 yes,
			"lambda.error":                  yes,
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
	case AWSLambdaInvokeSpanType:
		return map[string]struct{}{
			"function": yes,
			"type":     yes,
			"error":    yes,
		}
	case LogSpanType:
		return map[string]struct{}{
			"log.message":    yes,
			"log.level":      yes,
			"log.parameters": yes,
			"log.logger":     yes,
		}
	case MongoDBSpanType:
		return map[string]struct{}{
			"mongo.service":   yes,
			"mongo.namespace": yes,
			"mongo.command":   yes,
			"mongo.query":     yes,
			"mongo.json":      yes,
			"mongo.filter":    yes,
			"mongo.error":     yes,
		}
	case PostgreSQLSpanType:
		return map[string]struct{}{
			"pg.db":    yes,
			"pg.user":  yes,
			"pg.stmt":  yes,
			"pg.host":  yes,
			"pg.port":  yes,
			"pg.error": yes,
		}
	case RedisSpanType:
		return map[string]struct{}{
			"redis.connection":  yes,
			"redis.command":     yes,
			"redis.subCommands": yes,
			"redis.error":       yes,
		}
	case RabbitMQSpanType:
		return map[string]struct{}{
			"rabbitmq.exchange": yes,
			"rabbitmq.key":      yes,
			"rabbitmq.sort":     yes,
			"rabbitmq.address":  yes,
			"rabbitmq.error":    yes,
		}
	case AzureFunctionType:
		return map[string]struct{}{
			"azf.name":         yes,
			"azf.functionname": yes,
			"azf.methodname":   yes,
			"azf.triggername":  yes,
			"azf.runtime":      yes,
			"azf.error":        yes,
		}
	case GraphQLServerType, GraphQLClientType:
		return map[string]struct{}{
			"graphql.operationName": yes,
			"graphql.operationType": yes,
			"graphql.fields":        yes,
			"graphql.args":          yes,
			"graphql.error":         yes,
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

// newHTTPSpanData initializes a new HTTP span data from tracer span
func newHTTPSpanData(span *spanS) HTTPSpanData {
	data := HTTPSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     newHTTPSpanTags(span),
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
	// RouteID is an optional name/identifier for the matched route
	RouteID string `json:"route_id,omitempty"`
	// The name:port of the host to which the request had been sent
	Host string `json:"host,omitempty"`
	// The name of the protocol used for request ("http" or "https")
	Protocol string `json:"protocol,omitempty"`
	// The message describing an error occurred during the request handling
	Error string `json:"error,omitempty"`
}

// newHTTPSpanTags extracts HTTP-specific span tags from a tracer span
func newHTTPSpanTags(span *spanS) HTTPSpanTags {
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
		case "http.route_id":
			readStringTag(&tags.RouteID, v)
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

// newRPCSpanData initializes a new RPC span data from tracer span
func newRPCSpanData(span *spanS) RPCSpanData {
	data := RPCSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     newRPCSpanTags(span),
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

// newRPCSpanTags extracts RPC-specific span tags from a tracer span
func newRPCSpanTags(span *spanS) RPCSpanTags {
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

// newKafkaSpanData initializes a new Kafka span data from tracer span
func newKafkaSpanData(span *spanS) KafkaSpanData {
	data := KafkaSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     newKafkaSpanTags(span),
	}

	kindTag := span.Tags[string(ext.SpanKind)]
	data.producerSpan = kindTag == ext.SpanKindProducerEnum || kindTag == string(ext.SpanKindProducerEnum)

	return data
}

// Kind returns instana.ExitSpanKind for producer spans and instana.EntrySpanKind otherwise
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

// newKafkaSpanTags extracts Kafka-specific span tags from a tracer span
func newKafkaSpanTags(span *spanS) KafkaSpanTags {
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

// RabbitMQSpanData represents the `data` section of an RabbitMQ span
type RabbitMQSpanData struct {
	SpanData
	Tags RabbitMQSpanTags `json:"rabbitmq"`

	producerSpan bool
}

// newRabbitMQSpanData initializes a new RabbitMQ span data from tracer span
func newRabbitMQSpanData(span *spanS) RabbitMQSpanData {
	data := RabbitMQSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     newRabbitMQSpanTags(span),
	}

	kindTag := span.Tags[string(ext.SpanKind)]
	data.producerSpan = kindTag == ext.SpanKindProducerEnum || kindTag == string(ext.SpanKindProducerEnum)

	return data
}

// Kind returns instana.ExitSpanKind for producer spans and instana.EntrySpanKind otherwise
func (d RabbitMQSpanData) Kind() SpanKind {
	if d.producerSpan {
		return ExitSpanKind
	}

	return EntrySpanKind
}

// RabbitMQSpanTags contains fields within the `data.rabbitmq` section
type RabbitMQSpanTags struct {
	// The RabbitMQ exchange name
	Exchange string `json:"exchange"`
	// The routing key
	Key string `json:"key"`
	// Indicates wether the message is being produced or consumed
	Sort string `json:"sort"`
	// The AMQP URI used to establish a connection to RabbitMQ
	Address string `json:"address"`
	// Error is the optional error that can be thrown by RabbitMQ when executing a command
	Error string `json:"error,omitempty"`
}

// newRabbitMQSpanTags extracts RabbitMQ-specific span tags from a tracer span
func newRabbitMQSpanTags(span *spanS) RabbitMQSpanTags {
	var tags RabbitMQSpanTags
	for k, v := range span.Tags {
		switch k {
		case "rabbitmq.exchange":
			readStringTag(&tags.Exchange, v)
		case "rabbitmq.key":
			readStringTag(&tags.Key, v)
		case "rabbitmq.sort":
			readStringTag(&tags.Sort, v)
		case "rabbitmq.address":
			readStringTag(&tags.Address, v)
		case "rabbitmq.error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}

// GCPStorageSpanData represents the `data` section of a Google Cloud Storage span sent within an OT span document
type GCPStorageSpanData struct {
	SpanData
	Tags GCPStorageSpanTags `json:"gcs"`
}

// newGCPStorageSpanData initializes a new Google Cloud Storage span data from tracer span
func newGCPStorageSpanData(span *spanS) GCPStorageSpanData {
	data := GCPStorageSpanData{
		SpanData: NewSpanData(span, GCPStorageSpanType),
		Tags:     newGCPStorageSpanTags(span),
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

// newGCPStorageSpanTags extracts Google Cloud Storage span tags from a tracer span
func newGCPStorageSpanTags(span *spanS) GCPStorageSpanTags {
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

// newGCPPubSubSpanData initializes a new Google Cloud Pub/Span span data from tracer span
func newGCPPubSubSpanData(span *spanS) GCPPubSubSpanData {
	data := GCPPubSubSpanData{
		SpanData: NewSpanData(span, GCPPubSubSpanType),
		Tags:     newGCPPubSubSpanTags(span),
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

// newGCPPubSubSpanTags extracts Google Cloud Pub/Sub span tags from a tracer span
func newGCPPubSubSpanTags(span *spanS) GCPPubSubSpanTags {
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

// newAWSLambdaCloudWatchSpanTags extracts CloudWatch tags for an AWS Lambda entry span
func newAWSLambdaCloudWatchSpanTags(span *spanS) AWSLambdaCloudWatchSpanTags {
	var tags AWSLambdaCloudWatchSpanTags

	if events := newAWSLambdaCloudWatchEventTags(span); !events.IsZero() {
		tags.Events = &events
	}

	if logs := newAWSLambdaCloudWatchLogsTags(span); !logs.IsZero() {
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

// newAWSLambdaCloudWatchEventTags extracts CloudWatch event tags for an AWS Lambda entry span. It truncates
// the resources list to the first 3 items, populating the `data.lambda.cw.events.more` tag and limits each
// resource string to the first 200 characters to reduce the payload.
func newAWSLambdaCloudWatchEventTags(span *spanS) AWSLambdaCloudWatchEventTags {
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

// newAWSLambdaCloudWatchLogsTags extracts CloudWatch Logs tags for an AWS Lambda entry span. It truncates
// the log events list to the first 3 items, populating the `data.lambda.cw.logs.more` tag and limits each
// log string to the first 200 characters to reduce the payload.
func newAWSLambdaCloudWatchLogsTags(span *spanS) AWSLambdaCloudWatchLogsTags {
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

// newAWSLambdaS3SpanTags extracts S3 Event tags for an AWS Lambda entry span. It truncates
// the events list to the first 3 items and limits each object names to the first 200 characters to reduce the payload.
func newAWSLambdaS3SpanTags(span *spanS) AWSLambdaS3SpanTags {
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

// newAWSLambdaSQSSpanTags extracts SQS event tags for an AWS Lambda entry span. It truncates
// the events list to the first 3 items to reduce the payload.
func newAWSLambdaSQSSpanTags(span *spanS) AWSLambdaSQSSpanTags {
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
	// ColdStart is true if this is the first time current instance of the function was invoked
	ColdStart bool `json:"coldStart,omitempty"`
	// MillisecondsLeft is a number of milliseconds until timeout
	MillisecondsLeft int `json:"msleft,omitempty"`
	// Error is an AWS Lambda specific error
	Error string `json:"error,omitempty"`
	// CloudWatch holds the details of a CloudWatch event associated with this lambda
	CloudWatch *AWSLambdaCloudWatchSpanTags `json:"cw,omitempty"`
	// S3 holds the details of a S3 events associated with this lambda
	S3 *AWSLambdaS3SpanTags
	// SQS holds the details of a SQS events associated with this lambda
	SQS *AWSLambdaSQSSpanTags
}

// newAWSLambdaSpanTags extracts AWS Lambda entry span tags from a tracer span
func newAWSLambdaSpanTags(span *spanS) AWSLambdaSpanTags {
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

	if v, ok := span.Tags["lambda.coldStart"]; ok {
		readBoolTag(&tags.ColdStart, v)
	}

	if v, ok := span.Tags["lambda.msleft"]; ok {
		readIntTag(&tags.MillisecondsLeft, v)
	}

	if v, ok := span.Tags["lambda.error"]; ok {
		readStringTag(&tags.Error, v)
	}

	if cw := newAWSLambdaCloudWatchSpanTags(span); !cw.IsZero() {
		tags.CloudWatch = &cw
	}

	if st := newAWSLambdaS3SpanTags(span); !st.IsZero() {
		tags.S3 = &st
	}

	if sqs := newAWSLambdaSQSSpanTags(span); !sqs.IsZero() {
		tags.SQS = &sqs
	}

	return tags
}

// AWSLambdaSpanData is the base span data type for AWS Lambda entry spans
type AWSLambdaSpanData struct {
	Snapshot AWSLambdaSpanTags `json:"lambda"`
	HTTP     *HTTPSpanTags     `json:"http,omitempty"`
}

// newAWSLambdaSpanData initializes a new AWSLambdaSpanData from span
func newAWSLambdaSpanData(span *spanS) AWSLambdaSpanData {
	d := AWSLambdaSpanData{
		Snapshot: newAWSLambdaSpanTags(span),
	}

	switch span.Tags["lambda.trigger"] {
	case "aws:api.gateway", "aws:application.load.balancer":
		tags := newHTTPSpanTags(span)
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

// newAWSS3SpanData initializes a new AWS S3 span data from tracer span
func newAWSS3SpanData(span *spanS) AWSS3SpanData {
	data := AWSS3SpanData{
		SpanData: NewSpanData(span, AWSS3SpanType),
		Tags:     newAWSS3SpanTags(span),
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

// newAWSS3SpanTags extracts AWS S3 span tags from a tracer span
func newAWSS3SpanTags(span *spanS) AWSS3SpanTags {
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

// newAWSSQSSpanData initializes a new AWS SQS span data from tracer span
func newAWSSQSSpanData(span *spanS) AWSSQSSpanData {
	data := AWSSQSSpanData{
		SpanData: NewSpanData(span, AWSSQSSpanType),
		Tags:     newAWSSQSSpanTags(span),
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

// newAWSSQSSpanTags extracts AWS SQS span tags from a tracer span
func newAWSSQSSpanTags(span *spanS) AWSSQSSpanTags {
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

// newAWSSNSSpanData initializes a new AWS SNS span data from tracer span
func newAWSSNSSpanData(span *spanS) AWSSNSSpanData {
	data := AWSSNSSpanData{
		SpanData: NewSpanData(span, AWSSNSSpanType),
		Tags:     newAWSSNSSpanTags(span),
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

// newAWSSNSSpanTags extracts AWS SNS span tags from a tracer span
func newAWSSNSSpanTags(span *spanS) AWSSNSSpanTags {
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
	Tags AWSDynamoDBSpanTags `json:"dynamodb"`
}

// newAWSDynamoDBSpanData initializes a new AWS DynamoDB span data from tracer span
func newAWSDynamoDBSpanData(span *spanS) AWSDynamoDBSpanData {
	data := AWSDynamoDBSpanData{
		SpanData: NewSpanData(span, AWSDynamoDBSpanType),
		Tags:     newAWSDynamoDBSpanTags(span),
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
	// Region is a region from the AWS session config
	Region string `json:"region,omitempty"`
}

// newAWSDynamoDBSpanTags extracts AWS DynamoDB span tags from a tracer span
func newAWSDynamoDBSpanTags(span *spanS) AWSDynamoDBSpanTags {
	var tags AWSDynamoDBSpanTags
	for k, v := range span.Tags {
		switch k {
		case "dynamodb.table":
			readStringTag(&tags.Table, v)
		case "dynamodb.op":
			readStringTag(&tags.Operation, v)
		case "dynamodb.error":
			readStringTag(&tags.Error, v)
		case "dynamodb.region":
			readStringTag(&tags.Region, v)
		}
	}

	return tags
}

// AWSInvokeSpanTags contains fields within the `aws.lambda.invoke` section of an OT span document
type AWSInvokeSpanTags struct {
	// FunctionName is a name of the function which is invoked
	FunctionName string `json:"function"`
	// InvocationType if equal to `Event`, means it is an async invocation
	InvocationType string `json:"type"`
	// Error is an optional error returned by AWS API
	Error string `json:"error,omitempty"`
}

func newAWSDInvokeSpanTags(span *spanS) AWSInvokeSpanTags {
	var tags AWSInvokeSpanTags
	for k, v := range span.Tags {
		switch k {
		case "function":
			readStringTag(&tags.FunctionName, v)
		case "type":
			readStringTag(&tags.InvocationType, v)
		case "error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}

// AWSLambdaInvokeSpanData represents the `data` section of a AWS Invoke span sent within an OT span document
type AWSLambdaInvokeSpanData struct {
	SpanData
	Tags AWSInvokeSpanTags `json:"aws.lambda.invoke"`
}

// Kind returns the span kind for a AWS SDK Invoke span
func (d AWSLambdaInvokeSpanData) Kind() SpanKind {
	return ExitSpanKind
}

// Type returns the span type for an AWS SDK Invoke span
func (d AWSLambdaInvokeSpanData) Type() RegisteredSpanType {
	return AWSLambdaInvokeSpanType
}

// newAWSLambdaInvokeSpanData initializes a new AWS Invoke span data from tracer span
func newAWSLambdaInvokeSpanData(span *spanS) AWSLambdaInvokeSpanData {
	data := AWSLambdaInvokeSpanData{
		SpanData: NewSpanData(span, AWSLambdaInvokeSpanType),
		Tags:     newAWSDInvokeSpanTags(span),
	}

	return data
}

// LogSpanData represents the `data` section of a logging span
type LogSpanData struct {
	SpanData
	Tags LogSpanTags `json:"log"`
}

// newLogSpanData initializes a new logging span data from tracer span
func newLogSpanData(span *spanS) LogSpanData {
	return LogSpanData{
		SpanData: NewSpanData(span, LogSpanType),
		Tags:     newLogSpanTags(span),
	}
}

// Kind returns the span kind for a logging span
func (d LogSpanData) Kind() SpanKind {
	return ExitSpanKind
}

// LogSpanTags contains fields within the `data.log` section of an OT span document
type LogSpanTags struct {
	// Message is a string to log
	Message string `json:"message"`
	// Level is an optional log level for this record, e.g. INFO
	Level string `json:"level,omitempty"`
	// Logger is an optional logger name
	Logger string `json:"logger,omitempty"`
	// Error is an optional error string (if any)
	Error string `json:"parameters,omitempty"`
}

func newLogSpanTags(span *spanS) LogSpanTags {
	var tags LogSpanTags
	for k, v := range span.Tags {
		switch k {
		case "log.message":
			readStringTag(&tags.Message, v)
		case "log.level":
			readStringTag(&tags.Level, v)
		case "log.parameters":
			readStringTag(&tags.Error, v)
		case "log.logger":
			readStringTag(&tags.Logger, v)
		}
	}

	return tags
}

// MongoDBSpanData represents the `data` section of a MongoDB client span
type MongoDBSpanData struct {
	SpanData
	Tags MongoDBSpanTags `json:"mongo"`
}

// newMongoDBSpanData initializes a new MongoDB clientspan data from tracer span
func newMongoDBSpanData(span *spanS) MongoDBSpanData {
	return MongoDBSpanData{
		SpanData: NewSpanData(span, MongoDBSpanType),
		Tags:     newMongoDBSpanTags(span),
	}
}

// RedisSpanData represents the `data` section of a Redis client span
type RedisSpanData struct {
	SpanData
	Tags RedisSpanTags `json:"redis"`
}

// newRedisSpanData initializes a new Redis clientspan data from tracer span
func newRedisSpanData(span *spanS) RedisSpanData {
	return RedisSpanData{
		SpanData: NewSpanData(span, RedisSpanType),
		Tags:     newRedisSpanTags(span),
	}
}

// Kind returns the span kind for a Redis client span
func (d RedisSpanData) Kind() SpanKind {
	return ExitSpanKind
}

// Kind returns the span kind for a MongoDB client span
func (d MongoDBSpanData) Kind() SpanKind {
	return ExitSpanKind
}

// MongoDBSpanTags contains fields within the `data.mongo` section of an OT span document
type MongoDBSpanTags struct {
	// Service is the MongoDB server address in form of host:port
	Service string `json:"service"`
	// Namespace is the namespace name
	Namespace string `json:"namespace"`
	// Command is the name of the command initiated the span
	Command string `json:"command"`
	// Query is an optional query passed with command
	Query string `json:"query,omitempty"`
	// JSON is an optional JSON aggregation provided with command
	JSON string `json:"json,omitempty"`
	// Filter is an optional filter passed with command
	Filter string `json:"filter,omitempty"`
	// Error is an optional error message
	Error string `json:"error,omitempty"`
}

func newMongoDBSpanTags(span *spanS) MongoDBSpanTags {
	var tags MongoDBSpanTags
	for k, v := range span.Tags {
		switch k {
		case "mongo.service":
			readStringTag(&tags.Service, v)
		case "mongo.namespace":
			readStringTag(&tags.Namespace, v)
		case "mongo.command":
			readStringTag(&tags.Command, v)
		case "mongo.query":
			readStringTag(&tags.Query, v)
		case "mongo.json":
			readStringTag(&tags.JSON, v)
		case "mongo.filter":
			readStringTag(&tags.Filter, v)
		case "mongo.error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}

// PostgreSQLSpanData represents the `data` section of a PostgreSQL client span
type PostgreSQLSpanData struct {
	SpanData
	Tags postgreSQLSpanTags `json:"pg"`
}

// newPostgreSQLSpanData initializes a new PostgreSQL client span data from tracer span
func newPostgreSQLSpanData(span *spanS) PostgreSQLSpanData {
	return PostgreSQLSpanData{
		SpanData: NewSpanData(span, PostgreSQLSpanType),
		Tags:     newPostgreSQLSpanTags(span),
	}
}

// Kind returns the span kind for a PostgreSQL client span
func (d PostgreSQLSpanData) Kind() SpanKind {
	return ExitSpanKind
}

// postgreSQLSpanTags contains fields within the `data.pg` section of an OT span document
type postgreSQLSpanTags struct {
	Host string `json:"host"`
	Port string `json:"port"`
	DB   string `json:"db"`
	User string `json:"user"`
	Stmt string `json:"stmt"`

	Error string `json:"error,omitempty"`
}

func newPostgreSQLSpanTags(span *spanS) postgreSQLSpanTags {
	var tags postgreSQLSpanTags
	for k, v := range span.Tags {
		switch k {
		case "pg.host":
			readStringTag(&tags.Host, v)
		case "pg.port":
			readStringTag(&tags.Port, v)
		case "pg.db":
			readStringTag(&tags.DB, v)
		case "pg.stmt":
			readStringTag(&tags.Stmt, v)
		case "pg.user":
			readStringTag(&tags.User, v)
		case "pg.error":
		}
	}
	return tags
}

// RedisSpanTags contains fields within the `data.redis` section of an OT span document
type RedisSpanTags struct {
	// Connection is the host and port where the Redis server is running
	Connection string `json:"connection"`
	// Command is the Redis command being executed
	Command string `json:"command"`
	// Subcommands is the list of commands queued when a transaction starts, eg: by using the MULTI command
	Subcommands []string `json:"subCommands,omitempty"`
	// Error is the optional error that can be thrown by Redis when executing a command
	Error string `json:"error,omitempty"`
}

func newRedisSpanTags(span *spanS) RedisSpanTags {
	var tags RedisSpanTags
	for k, v := range span.Tags {
		switch k {
		case "redis.connection":
			readStringTag(&tags.Connection, v)
		case "redis.command":
			readStringTag(&tags.Command, v)
		case "redis.subCommands":
			readArrayStringTag(&tags.Subcommands, v)
		case "redis.error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}

type AZFSpanTags struct {
	Name         string `json:"name,omitempty"`
	FunctionName string `json:"functionname,omitempty"`
	MethodName   string `json:"methodname,omitempty"`
	Trigger      string `json:"triggername,omitempty"`
	Runtime      string `json:"runtime,omitempty"`
	Error        string `json:"error,omitempty"`
}

func newAZFSpanTags(span *spanS) AZFSpanTags {
	var tags AZFSpanTags
	for k, v := range span.Tags {
		switch k {
		case "azf.name":
			readStringTag(&tags.Name, v)
		case "azf.functionname":
			readStringTag(&tags.FunctionName, v)
		case "azf.methodname":
			readStringTag(&tags.MethodName, v)
		case "azf.triggername":
			readStringTag(&tags.Trigger, v)
		case "azf.runtime":
			readStringTag(&tags.Runtime, v)
		}
	}

	return tags
}

type AZFSpanData struct {
	SpanData
	Tags AZFSpanTags `json:"azf"`
}

func newAZFSpanData(span *spanS) AZFSpanData {
	return AZFSpanData{
		SpanData: NewSpanData(span, AzureFunctionType),
		Tags:     newAZFSpanTags(span),
	}
}

// Kind returns instana.EntrySpanKind for server spans and instana.ExitSpanKind otherwise
func (d AZFSpanData) Kind() SpanKind {
	return EntrySpanKind
}

// GraphQLSpanData represents the `data` section of a GraphQL span sent within an OT span document
type GraphQLSpanData struct {
	SpanData
	Tags GraphQLSpanTags `json:"graphql"`

	clientSpan bool
}

// newGraphQLSpanData initializes a new GraphQL span data from tracer span
func newGraphQLSpanData(span *spanS) GraphQLSpanData {
	data := GraphQLSpanData{
		SpanData: NewSpanData(span, RegisteredSpanType(span.Operation)),
		Tags:     newGraphQLSpanTags(span),
	}

	kindTag := span.Tags[string(ext.SpanKind)]
	data.clientSpan = kindTag == ext.SpanKindRPCClientEnum || kindTag == string(ext.SpanKindRPCClientEnum)

	return data
}

// Kind returns instana.EntrySpanKind for server spans and instana.ExitSpanKind otherwise
func (d GraphQLSpanData) Kind() SpanKind {
	if d.clientSpan {
		return ExitSpanKind
	}

	return EntrySpanKind
}

// GraphQLSpanTags contains fields within the `data.graphql` section of an OT span document
type GraphQLSpanTags struct {
	OperationName string              `json:"operationName,omitempty"`
	OperationType string              `json:"operationType,omitempty"`
	Fields        map[string][]string `json:"fields,omitempty"`
	Args          map[string][]string `json:"args,omitempty"`
	Error         string              `json:"error,omitempty"`
}

// newGraphQLSpanTags extracts GraphQL-specific span tags from a tracer span
func newGraphQLSpanTags(span *spanS) GraphQLSpanTags {
	var tags GraphQLSpanTags
	for k, v := range span.Tags {
		switch k {
		case "graphql.operationName":
			readStringTag(&tags.OperationName, v)
		case "graphql.operationType":
			readStringTag(&tags.OperationType, v)
		case "graphql.fields":
			readMapOfStringSlicesTag(&tags.Fields, v)
		case "graphql.args":
			readMapOfStringSlicesTag(&tags.Args, v)
		case "graphql.error":
			readStringTag(&tags.Error, v)
		}
	}

	return tags
}
