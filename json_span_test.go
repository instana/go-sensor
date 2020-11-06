package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/testify/assert"
	"github.com/opentracing/opentracing-go"
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
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

			sp := tracer.StartSpan(example.Operation)
			sp.Finish()

			spans := recorder.GetQueuedSpans()
			assert.Equal(t, 1, len(spans))
			span := spans[0]

			assert.IsType(t, example.Expected, span.Data)
		})
	}
}

func TestSpanKind_String(t *testing.T) {
	examples := map[string]struct {
		Kind     instana.SpanKind
		Expected string
	}{
		"entry": {
			Kind:     instana.EntrySpanKind,
			Expected: "entry",
		},
		"exit": {
			Kind:     instana.ExitSpanKind,
			Expected: "exit",
		},
		"intermediate": {
			Kind:     instana.IntermediateSpanKind,
			Expected: "intermediate",
		},
		"unknown": {
			Kind:     instana.SpanKind(0),
			Expected: "intermediate",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, example.Kind.String())
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
					ARN:     "lambda-arn-1",
					Runtime: "go",
					Name:    "test-lambda",
					Version: "42",
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
	}

	// ALB tags set is the same as for API Gateway, so we just copy it over
	examples["aws:application.load.balancer"] = examples["aws:api.gateway"]

	for trigger, example := range examples {
		t.Run(trigger, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

			sp := tracer.StartSpan("aws.lambda.entry", opentracing.Tags{
				"lambda.arn":     "lambda-arn-1",
				"lambda.name":    "test-lambda",
				"lambda.version": "42",
				"lambda.trigger": trigger,
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
