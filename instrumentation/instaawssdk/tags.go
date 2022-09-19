// (c) Copyright IBM Corp. 2022

package instaawssdk

// generic
const (
	errorTag = "error"
	typeTag  = "type"
)

// dynamodb
const (
	dynamodbOp     = "dynamodb.op"
	dynamodbTable  = "dynamodb.table"
	dynamodbError  = "dynamodb.error"
	dynamodbRegion = "dynamodb.region"
)

// lambda
const lambdaFunction = "function"

// s3
const (
	s3Region = "s3.region"
	s3Error  = "s3.error"
	s3Op     = "s3.op"
	s3Bucket = "s3.bucket"
	s3Key    = "s3.key"
)

// sns
const (
	snsError   = "sns.error"
	snsTopic   = "sns.topic"
	snsTarget  = "sns.target"
	snsPhone   = "sns.phone"
	snsSubject = "sns.subject"
)

// sqs
const (
	sqsSort  = "sqs.sort"
	sqsError = "sqs.error"
	sqsSize  = "sqs.size"
	sqsQueue = "sqs.queue"
	sqsType  = "sqs.type"
	sqsGroup = "sqs.group"
)
