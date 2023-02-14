// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instana_test

import (
	"errors"
	"strings"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestRegisteredSpanType_ExtractData(t *testing.T) {
	examples := map[string]struct {
		Operation string
		Expected  interface{}
	}{
		"net/http.Server": {
			Operation: "g.http",
			Expected:  instana.HTTPSpanData{},
		},
		"net/http.Client": {
			Operation: "http",
			Expected:  instana.HTTPSpanData{},
		},
		"golang.google.org/gppc.Server": {
			Operation: "rpc-server",
			Expected:  instana.RPCSpanData{},
		},
		"github.com/Shopify/sarama": {
			Operation: "kafka",
			Expected:  instana.KafkaSpanData{},
		},
		"sdk": {
			Operation: "test",
			Expected:  instana.SDKSpanData{},
		},
		"aws lambda": {
			Operation: "aws.lambda.entry",
			Expected:  instana.AWSLambdaSpanData{},
		},
		"aws s3": {
			Operation: "s3",
			Expected:  instana.AWSS3SpanData{},
		},
		"aws sqs": {
			Operation: "sqs",
			Expected:  instana.AWSSQSSpanData{},
		},
		"aws sns": {
			Operation: "sns",
			Expected:  instana.AWSSNSSpanData{},
		},
		"aws dynamodb": {
			Operation: "dynamodb",
			Expected:  instana.AWSDynamoDBSpanData{},
		},
		"aws invoke": {
			Operation: "aws.lambda.invoke",
			Expected:  instana.AWSLambdaInvokeSpanData{},
		},
		"logger": {
			Operation: "log.go",
			Expected:  instana.LogSpanData{},
		},
		"mongodb": {
			Operation: "mongo",
			Expected:  instana.MongoDBSpanData{},
		},
		"postgresql": {
			Operation: "postgres",
			Expected:  instana.PostgreSQLSpanData{},
		},
		"redis": {
			Operation: "redis",
			Expected:  instana.RedisSpanData{},
		},
		"rabbitmq": {
			Operation: "rabbitmq",
			Expected:  instana.RabbitMQSpanData{},
		},
		"graphql server": {
			Operation: "graphql.server",
			Expected:  instana.GraphQLSpanData{},
		},
		"graphql client": {
			Operation: "graphql.client",
			Expected:  instana.GraphQLSpanData{},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
			defer instana.ShutdownSensor()

			sp := tracer.StartSpan(example.Operation)
			sp.Finish()

			spans := recorder.GetQueuedSpans()
			assert.Equal(t, 1, len(spans))
			span := spans[0]

			assert.IsType(t, example.Expected, span.Data)
		})
	}
}

func TestNewAWSLambdaSpanData(t *testing.T) {
	examples := map[string]struct {
		Tags     opentracing.Tags
		Expected instana.AWSLambdaSpanData
	}{
		"aws:api.gateway": {
			Tags: opentracing.Tags{
				"http.protocol": "https",
				"http.url":      "https://example.com/lambda",
				"http.host":     "example.com",
				"http.method":   "GET",
				"http.path":     "/lambda",
				"http.params":   "q=test&secret=classified",
				"http.header":   map[string]string{"x-custom-header-1": "test"},
				"http.status":   404,
				"http.error":    "Not Found",
			},
			Expected: instana.AWSLambdaSpanData{
				Snapshot: instana.AWSLambdaSpanTags{
					ARN:              "lambda-arn-1",
					Runtime:          "go",
					Name:             "test-lambda",
					Version:          "42",
					Trigger:          "aws:api.gateway",
					ColdStart:        true,
					MillisecondsLeft: 5,
					Error:            "Not Found",
				},
				HTTP: &instana.HTTPSpanTags{
					URL:      "https://example.com/lambda",
					Status:   404,
					Method:   "GET",
					Path:     "/lambda",
					Params:   "q=test&secret=classified",
					Headers:  map[string]string{"x-custom-header-1": "test"},
					Host:     "example.com",
					Protocol: "https",
					Error:    "Not Found",
				},
			},
		},
		"aws:application.load.balancer": {
			Tags: opentracing.Tags{
				"http.protocol": "https",
				"http.url":      "https://example.com/lambda",
				"http.host":     "example.com",
				"http.method":   "GET",
				"http.path":     "/lambda",
				"http.params":   "q=test&secret=classified",
				"http.header":   map[string]string{"x-custom-header-1": "test"},
				"http.status":   404,
				"http.error":    "Not Found",
			},
			Expected: instana.AWSLambdaSpanData{
				Snapshot: instana.AWSLambdaSpanTags{
					ARN:              "lambda-arn-1",
					Runtime:          "go",
					Name:             "test-lambda",
					Version:          "42",
					Trigger:          "aws:application.load.balancer",
					ColdStart:        true,
					MillisecondsLeft: 5,
					Error:            "Not Found",
				},
				HTTP: &instana.HTTPSpanTags{
					URL:      "https://example.com/lambda",
					Status:   404,
					Method:   "GET",
					Path:     "/lambda",
					Params:   "q=test&secret=classified",
					Headers:  map[string]string{"x-custom-header-1": "test"},
					Host:     "example.com",
					Protocol: "https",
					Error:    "Not Found",
				},
			},
		},
		"aws:cloudwatch.events": {
			Tags: opentracing.Tags{
				"cloudwatch.events.id": "cw-event-1",
				"cloudwatch.events.resources": []string{
					"res1",
					strings.Repeat("long ", 40) + "res2",
					"res3",
					"res4",
				},
			},
			Expected: instana.AWSLambdaSpanData{
				Snapshot: instana.AWSLambdaSpanTags{
					ARN:              "lambda-arn-1",
					Runtime:          "go",
					Name:             "test-lambda",
					Version:          "42",
					Trigger:          "aws:cloudwatch.events",
					ColdStart:        true,
					MillisecondsLeft: 5,
					Error:            "Not Found",
					CloudWatch: &instana.AWSLambdaCloudWatchSpanTags{
						Events: &instana.AWSLambdaCloudWatchEventTags{
							ID: "cw-event-1",
							Resources: []string{
								"res1",
								strings.Repeat("long ", 40),
								"res3",
							},
							More: true,
						},
					},
				},
			},
		},
		"aws:cloudwatch.logs": {
			Tags: opentracing.Tags{
				"cloudwatch.logs.group":  "cw-log-group-1",
				"cloudwatch.logs.stream": "cw-log-stream-1",
				"cloudwatch.logs.events": []string{
					"log1",
					strings.Repeat("long ", 40) + "log2",
					"log3",
					"log4",
				},
				"cloudwatch.logs.decodingError": errors.New("none"),
			},
			Expected: instana.AWSLambdaSpanData{
				Snapshot: instana.AWSLambdaSpanTags{
					ARN:              "lambda-arn-1",
					Runtime:          "go",
					Name:             "test-lambda",
					Version:          "42",
					Trigger:          "aws:cloudwatch.logs",
					ColdStart:        true,
					MillisecondsLeft: 5,
					Error:            "Not Found",
					CloudWatch: &instana.AWSLambdaCloudWatchSpanTags{
						Logs: &instana.AWSLambdaCloudWatchLogsTags{
							Group:  "cw-log-group-1",
							Stream: "cw-log-stream-1",
							Events: []string{
								"log1",
								strings.Repeat("long ", 40),
								"log3",
							},
							More:          true,
							DecodingError: "none",
						},
					},
				},
			},
		},
		"aws:s3": {
			Tags: opentracing.Tags{
				"s3.events": []instana.AWSS3EventTags{
					{Name: "event1", Bucket: "bucket1", Object: "object1"},
					{Name: "event2", Bucket: "bucket2", Object: strings.Repeat("long ", 40) + "object2"},
					{Name: "event3", Bucket: "bucket3"},
					{Name: "event4", Bucket: "bucket4"},
				},
			},
			Expected: instana.AWSLambdaSpanData{
				Snapshot: instana.AWSLambdaSpanTags{
					ARN:              "lambda-arn-1",
					Runtime:          "go",
					Name:             "test-lambda",
					Version:          "42",
					Trigger:          "aws:s3",
					ColdStart:        true,
					MillisecondsLeft: 5,
					Error:            "Not Found",
					S3: &instana.AWSLambdaS3SpanTags{
						Events: []instana.AWSS3EventTags{
							{Name: "event1", Bucket: "bucket1", Object: "object1"},
							{Name: "event2", Bucket: "bucket2", Object: strings.Repeat("long ", 40)},
							{Name: "event3", Bucket: "bucket3"},
						},
					},
				},
			},
		},
		"aws:sqs": {
			Tags: opentracing.Tags{
				"sqs.messages": []instana.AWSSQSMessageTags{{Queue: "q1"}, {Queue: "q2"}, {Queue: "q3"}, {Queue: "q4"}},
			},
			Expected: instana.AWSLambdaSpanData{
				Snapshot: instana.AWSLambdaSpanTags{
					ARN:              "lambda-arn-1",
					Runtime:          "go",
					Name:             "test-lambda",
					Version:          "42",
					Trigger:          "aws:sqs",
					ColdStart:        true,
					MillisecondsLeft: 5,
					Error:            "Not Found",
					SQS: &instana.AWSLambdaSQSSpanTags{
						Messages: []instana.AWSSQSMessageTags{{Queue: "q1"}, {Queue: "q2"}, {Queue: "q3"}},
					},
				},
			},
		},
	}

	for trigger, example := range examples {
		t.Run(trigger, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
			defer instana.ShutdownSensor()

			sp := tracer.StartSpan("aws.lambda.entry", opentracing.Tags{
				"lambda.arn":       "lambda-arn-1",
				"lambda.name":      "test-lambda",
				"lambda.version":   "42",
				"lambda.trigger":   trigger,
				"lambda.coldStart": true,
				"lambda.msleft":    5,
				"lambda.error":     "Not Found",
			}, example.Tags)
			sp.Finish()

			spans := recorder.GetQueuedSpans()
			assert.Equal(t, 1, len(spans))
			span := spans[0]

			assert.Equal(t, "aws.lambda.entry", span.Name)
			assert.Equal(t, example.Expected, span.Data)
		})
	}
}
