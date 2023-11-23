# Instana instrumentation of AWS SDK v2 for Go

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]

This module contains the code for instrumenting AWS APIs which are based on the [`aws sdk v2`][aws-sdk-go-v2-github] library for Go. The following services are currently instrumented:

* [s3](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3)
* [DynamoDB](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb)
* [SQS](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/sqs)
* [SNS](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/sns)
* [Lambda](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/lambda)

## Installation

```bash
$ go get github.com/instana/go-sensor/instrumentation/instaawsv2
```

## Usage
To trace the AWS APIs, the user has to instrument the [`aws.Config`][aws-config] object using the [`instaawsv2.Instrument`][Instrument] function before starting to use the AWS service clients like s3, dynamodb etc.

## Special cases

### 1. Instrumenting SQS consumers
An SQS client that uses instrumented `aws.Config` automatically creates entry spans for each incoming
`sqs.Message`. To use this entry span context as a parent in your message handler use
[`instaawsv2.SpanContextFromSQSMessage`][SpanContextFromSQSMessage]:

```go
func handleMessage(ctx context.Context, msg *sqs.Message) {
	if parent, ok := instaawsv2.SpanContextFromSQSMessage(msg, tracer); ok {
		sp := tracer.StartSpan("handleMessage", opentracing.ChildOf(parent))
		defer sp.Finish()

		ctx = instana.ContextWithSpan(ctx, sp)
    }

    // ...
}
```

### 2. Instrumenting calls to AWS Lambda
When calls to AWS Lambda is instrumented using instaawsv2, the trace context is propagated inside a `ClientContext.Custom` field in the `InvokeInput` object. The reserved keys used for this are: 
1. `x-instana-t`
2. `x-instana-s`
3. `x-instana-l`

Hence, to avoid collisions, it is recommended to avoid these keys in your application code. 

## Limitations
- Current instrumentation does not support asynchronous lambda invocation.
- If the length of base64 encoded ClientContext will exceed 3582 bytes, tracing headers will be not propagated.


[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaawsv2
[Instrument]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaawsv2?tab=doc#Instrument
[SpanContextFromSQSMessage]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaawsv2?tab=doc#SpanContextFromSQSMessage
[aws-sdk-go-v2-github]: https://github.com/aws/aws-sdk-go-v2
[aws-config]: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/config#Config

<!---
Mandatory comment section for CI/CD !!
target-pkg-url: github.com/aws/aws-sdk-go-v2
current-version: v1.21.0
--->
