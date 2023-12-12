module github.com/instana/go-sensor/instrumentation/instacosmos

go 1.18

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.9.0
	github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos v0.3.6
	github.com/google/uuid v1.4.0
	github.com/instana/go-sensor v1.58.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/instana/go-sensor => ../../
