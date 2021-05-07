Instana instrumentation for AWS SDK for Go v1
=============================================

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]

This module contains instrumentation code for AWS API clients that use `github.com/aws/aws-sdk-go` library `v1.8.0` and above.

Following services are currently instrumented:

* [DynamoDB](https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/)
* [S3](https://docs.aws.amazon.com/sdk-for-go/api/service/s3/)
* [SNS](https://docs.aws.amazon.com/sdk-for-go/api/service/sns/)
* [SQS](https://docs.aws.amazon.com/sdk-for-go/api/service/sqs/)
* [Lambda](https://docs.aws.amazon.com/sdk-for-go/api/service/lambda/) 
  - Read about usage and limitations [here](https://github.com/instana/go-sensor/tree/master/instrumentation/instaawssdk#instrumenting-lambda)

Installation
------------

```bash
$ go get github.com/instana/go-sensor/instrumentation/instaawssdk
```

Usage
-----

This instrumentation requires an [`instana.Sensor`][Sensor] to initialize spans and handle the trace context propagation.
You can create a new instance of Instana tracer using [`instana.NewSensor()`][NewSensor].

To trace requests made to the AWS API instrument the `aws/session.Session` using [`instaawssdk.InstrumentSession()`][InstrumentSession]
before creating the service client:

```go
sess := session.Must(session.NewSession(&aws.Config{}))

// Initialize Instana sensor
sensor := instana.NewSensor("my-aws-app")
// Instrument aws/session.Session
instaawssdk.InstrumentSession(sess, sensor)

// Create a service client using instrumented session
dynamoDBClient := dynamodb.New(sess)

// Use service client as usual
// ...
```

Instana tracer uses `context.Context` to propagate the trace context. To ensure trace continuation within
the instrumented service **use AWS SDK client methods that take `context.Context` as an argument**.
Usually these method names end with `WithContext` suffix, e.g.

* `(*dynamodb.Client).PutItemWithContext()`
* `(*s3.Client).CreateBucketWithContext()`
* `(*sns.Client).PublishWithContext()`
* `(*sqs.Client).ReceiveMessagesWithContext()`
* `(*lambda.Lambda).InvokeWithContext()`
* etc.

### Instrumenting SQS consumers

An SQS client that uses instrumented `session.Session` automatically creates entry spans for each incoming
`sqs.Message`. To use this entry span context as a parent in your message handler use
[`instaawssdk.SpanContextFromSQSMessage()`][SpanContextFromSQSMessage]:

```go
func handleMessage(ctx context.Context, msg *sqs.Message) {
	if parent, ok := instaawssdk.SpanContextFromSQSMessage(msg, sensor); ok {
		sp := sensor.Tracer().StartSpan("handleMessage", opentracing.ChildOf(parent))
		defer sp.Finish()

		ctx = instana.ContextWithSpan(ctx, sp)
    }

    // ...
}
```

### Instrumenting lambda

If a session is instrumented, it will propagate tracing context automatically using values from the `ctx`.

Example:
```go
sensor := instana.NewSensor("my-new-sensor")
sess, _ := session.NewSession()
instaawssdk.InstrumentSession(sess, sensor)
svc := sdk.New(sess)
input := &sdk.InvokeInput{
    FunctionName: "my-lambda-function-name",
    // this field is optional
    // IMPORTANT type `Event` is not supported by the instrumentation
    InvocationType: aws.String("RequestResponse"), 
    Payload: []byte("{}"),
}

// invoke with context, otherwise, you will need to set context manually to propagate tracing data
svc.InvokeWithContext(ctx, input)
```

Tracing context propagated inside a `ClientContext.Custom` field in the `InvokeInput` object. Reserved keys are:
- `X-INSTANA-T`
- `X-INSTANA-S`
- `X-INSTANA-L`
- `TRACEPARENT`
- `TRACESTATE`

To avoid a collision, do not set them in your application code.

Known limitation:
- Current instrumentation does not support asynchronous lambda invocation.
- If the length of base64 encoded `ClientContext` will exceed 3582 bytes, tracing headers will be not propagated.
- Deprecated methods like `InvokeAsync`, `InvokeAsyncWithContext` etc. are not supported.

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaawssdk
[Sensor]: https://pkg.go.dev/github.com/instana/go-sensor?tab=doc#Sensor
[NewSensor]: https://pkg.go.dev/github.com/instana/go-sensor?tab=doc#NewSensor
[InstrumentSession]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaawssdk?tab=doc#InstrumentSession
[SpanContextFromSQSMessage]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaawssdk?tab=doc#SpanContextFromSQSMessage
