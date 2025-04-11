module cosmos.example

go 1.23.0

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.16.0
	github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos v1.3.0
	github.com/google/uuid v1.6.0
	github.com/instana/go-sensor v1.67.1
	github.com/instana/go-sensor/instrumentation/instacosmos v1.12.2
)

require (
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	golang.org/x/net v0.36.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace (
	github.com/instana/go-sensor => ../../
	github.com/instana/go-sensor/instrumentation/instacosmos v1.12.2 => ../../instrumentation/instacosmos
)
