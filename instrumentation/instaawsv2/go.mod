module github.com/instana/go-sensor/instrumentation/instaawsv2

go 1.15

require (
	github.com/aws/aws-sdk-go-v2 v1.23.1
	github.com/aws/aws-sdk-go-v2/config v1.18.40
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.10.39
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.21.5
	github.com/aws/aws-sdk-go-v2/service/lambda v1.39.5
	github.com/aws/aws-sdk-go-v2/service/rds v1.54.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.38.5
	github.com/aws/aws-sdk-go-v2/service/sns v1.22.0
	github.com/aws/aws-sdk-go-v2/service/sqs v1.24.5
	github.com/aws/smithy-go v1.17.0
	github.com/instana/go-sensor v1.58.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.1
)
