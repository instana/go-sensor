// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/aws/aws-sdk-go/aws/request"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
)

// StartInvokeLambdaSpan initiates a new span from an AWS Invoke request and injects it into the
// request.Request context
func StartInvokeLambdaSpan(req *request.Request, sensor instana.TracerLogger) {

	// an exit span will be created without a parent span
	// and forwarded if user chose to opt in
	opts := []opentracing.StartSpanOption{}
	parent, ok := instana.SpanFromContext(req.Context())
	if ok {
		opts = append(opts, opentracing.ChildOf(parent.Context()))
	}
	sp := sensor.Tracer().StartSpan("aws.lambda.invoke", opts...)

	if ii, ok := req.Params.(*lambda.InvokeInput); ok {
		sp.SetTag(lambdaFunction, aws.StringValue(ii.FunctionName))

		invocationType := aws.StringValue(ii.InvocationType)
		if invocationType == "" {
			invocationType = lambda.InvocationTypeRequestResponse
		}
		sp.SetTag(typeTag, invocationType)
	}

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
	injectTraceContext(sp, req, sensor.Logger())
}

// FinalizeInvokeLambdaSpan retrieves error from completed request.Request if any and adds it
// to the span
func FinalizeInvokeLambdaSpan(req *request.Request) {
	sp, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}
	defer sp.Finish()

	if req.Error != nil {
		sp.LogFields(otlog.Error(req.Error))
		sp.SetTag(errorTag, req.Error.Error())
	}
}
