package instana

type spanCategory string

const (
	httpReq   spanCategory = "http"
	rpc       spanCategory = "rpc"
	messaging spanCategory = "messaging"
	logging   spanCategory = "logging"
	databases spanCategory = "databases"
	graphql   spanCategory = "graphql"
	unknown   spanCategory = "unknown"
)

func (c spanCategory) String() string {
	return string(c)
}

func (opts *TracerOptions) DisableAllCategories() {
	opts.Disable = map[string]bool{
		httpReq.String():   true,
		rpc.String():       true,
		messaging.String(): true,
		logging.String():   true,
		databases.String(): true,
		graphql.String():   true,
	}
}

var genericSpanMap = map[string]spanCategory{
	"sdk.database": databases,
}

// Not categorized spans
// GCPStorageSpanType = RegisteredSpanType("gcs")
// AWSLambdaEntrySpanType = RegisteredSpanType("aws.lambda.entry")
// AWSSQSSpanType = RegisteredSpanType("sqs")
// AWSSNSSpanType = RegisteredSpanType("sns")
// AWSLambdaInvokeSpanType = RegisteredSpanType("aws.lambda.invoke")
// AzureFunctionType = RegisteredSpanType("azf")

var registeredSpanMap map[RegisteredSpanType]spanCategory = map[RegisteredSpanType]spanCategory{

	// http
	HTTPServerSpanType: httpReq,
	HTTPClientSpanType: httpReq,

	// rpc
	RPCServerSpanType: rpc,
	RPCClientSpanType: rpc,

	// messaging
	KafkaSpanType:     messaging,
	GCPPubSubSpanType: messaging,
	RabbitMQSpanType:  messaging,

	// logging
	LogSpanType: logging,

	// databases
	AWSDynamoDBSpanType: databases,
	AWSS3SpanType:       databases,
	MongoDBSpanType:     databases,
	PostgreSQLSpanType:  databases,
	MySQLSpanType:       databases,
	RedisSpanType:       databases,
	CouchbaseSpanType:   databases,
	CosmosSpanType:      databases,

	// graphql
	GraphQLServerType: graphql,
	GraphQLClientType: graphql,
}

func (r *spanS) getSpanCategory() spanCategory {
	// return span category if it is a registered span type
	if c, ok := registeredSpanMap[RegisteredSpanType(r.Operation)]; ok {
		return c
	}

	// return span category if it is a generic span type
	if c, ok := genericSpanMap[r.Operation]; ok {
		return c
	}

	return unknown
}

func (c spanCategory) Enabled() bool {
	return !sensor.options.Tracer.Disable[c.String()]
}
