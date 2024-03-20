module cosmos.example

go 1.21

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.9.0
	github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos v0.3.6
	github.com/google/uuid v1.4.0
	github.com/instana/go-sensor v1.60.0
	github.com/instana/go-sensor/instrumentation/instacosmos v1.0.0
)

require (
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.0 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	github.com/instana/go-sensor => ../../
	github.com/instana/go-sensor/instrumentation/instacosmos v1.0.0 => ../../instrumentation/instacosmos
)
