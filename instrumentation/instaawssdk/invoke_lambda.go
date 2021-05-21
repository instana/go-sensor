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
func StartInvokeLambdaSpan(req *request.Request, sensor *instana.Sensor) {
	parent, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}

	sp := sensor.Tracer().StartSpan("aws.lambda.invoke",
		opentracing.ChildOf(parent.Context()),
	)

	if ii, ok := req.Params.(*lambda.InvokeInput); ok {
		sp.SetTag("function", aws.StringValue(ii.FunctionName))

		invocationType := aws.StringValue(ii.InvocationType)
		if invocationType == "" {
			invocationType = lambda.InvocationTypeRequestResponse
		}
		sp.SetTag("type", invocationType)
	}

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
	injectTraceContext(sensor.Logger(), sp, req)
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
		sp.SetTag("error", req.Error.Error())
	}
}
