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

// StartSNSSpan initiates a new span from an AWS SNS request and injects it into the
// request.Request context
func StartSNSSpan(req *request.Request, sensor *instana.Sensor) {
	parent, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}

	sp := sensor.Tracer().StartSpan("sns",
		ext.SpanKindRPCClient,
		opentracing.ChildOf(parent.Context()),
	)

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
}

// FinalizeSNSSpan retrieves tags from completed request.Request and adds them
// to the span
func FinalizeSNSSpan(req *request.Request) {
	sp, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}
	defer sp.Finish()

	if req.Error != nil {
		sp.LogFields(otlog.Error(req.Error))
		sp.SetTag("sns.error", req.Error.Error())
	}
}
