package instaawssdk

import (
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/lambda"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func StartInvokeLambdaSpan(req *request.Request, sensor *instana.Sensor) {
	tags := opentracing.Tags{}
	if ii, ok := req.Params.(*lambda.InvokeInput); ok {
		if ii.FunctionName != nil {
			tags["invoke.function"] = *ii.FunctionName
		}

		if ii.InvocationType != nil {
			tags["invoke.type"] = *ii.InvocationType
		}
	}

	parent, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}

	sp := sensor.Tracer().StartSpan("aws.sdk.invoke",
		ext.SpanKindRPCClient,
		opentracing.ChildOf(parent.Context()),
		tags,
	)

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
	injectTraceContext(sp, req)
}

func FinalizeInvokeLambdaSpan(req *request.Request) {
	sp, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}
	defer sp.Finish()

	if req.Error != nil {
		sp.LogFields(otlog.Error(req.Error))
		sp.SetTag("invoke.error", req.Error.Error())
	}
}
