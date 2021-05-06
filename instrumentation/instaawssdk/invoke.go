package instaawssdk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/lambda"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func StartInvokeSpan(req *request.Request, sensor *instana.Sensor) {
	tags := opentracing.Tags{}
	if ii, ok := req.Params.(lambda.InvokeInput); ok {
		tags["invoke.function"] = *ii.FunctionName
		tags["invoke.type"] = *ii.InvocationType
	}

	parent, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}

	sp := sensor.Tracer().StartSpan("invoke",
		ext.SpanKindRPCClient,
		opentracing.ChildOf(parent.Context()),
		tags,
	)

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
	injectTraceContext(sp, req)
}

func FinalizeInvokeSpan(req *request.Request) {
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

func encodeToBase64(v interface{}) (string, error) {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	err := json.NewEncoder(encoder).Encode(v)
	if err != nil {
		return "", err
	}
	encoder.Close()
	return buf.String(), nil
}
