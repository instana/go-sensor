// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"github.com/aws/aws-sdk-go/aws/request"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// StartDynamoDBSpan initiates a new span from an AWS DynamoDB request and
// injects it into the request.Request context
func StartDynamoDBSpan(req *request.Request, sensor instana.TracerLogger) {
	tags, err := extractDynamoDBTags(req)
	if err != nil {
		if err == errMethodNotInstrumented {
			return
		}

		sensor.Logger().Warn("failed to extract DynamoDB tags: ", err)
	}

	// an exit span will be created without a parent span
	// and forwarded if user chose to opt in
	opts := []opentracing.StartSpanOption{
		ext.SpanKindRPCClient,
		tags,
	}
	parent, ok := instana.SpanFromContext(req.Context())
	if ok {
		opts = append(opts, opentracing.ChildOf(parent.Context()))
	}
	sp := sensor.Tracer().StartSpan("dynamodb", opts...)

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
}

// FinalizeDynamoDBSpan retrieves tags from completed request.Request and adds them
// to the span
func FinalizeDynamoDBSpan(req *request.Request) {
	sp, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}
	defer sp.Finish()

	if req.Error != nil {
		sp.LogFields(otlog.Error(req.Error))
		sp.SetTag(dynamodbError, req.Error.Error())
	}
}
