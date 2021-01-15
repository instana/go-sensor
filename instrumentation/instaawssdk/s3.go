// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"github.com/aws/aws-sdk-go/aws/request"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// StartS3Span initiates a new span from AWS S3 request and injects it into the
// request.Request context
func StartS3Span(req *request.Request, sensor *instana.Sensor) {
	parent, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}

	sp := sensor.Tracer().StartSpan("s3",
		ext.SpanKindRPCClient,
		opentracing.ChildOf(parent.Context()),
		opentracing.Tags{
			"s3.region": req.ClientInfo.SigningRegion,
			"s3.op":     req.Operation.Name,
		},
	)

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
}

// FinalizeS3Span retrieves tags from completed request.Request and adds them
// to the span
func FinalizeS3Span(sp opentracing.Span, req *request.Request) {
	if req.Error != nil {
		sp.SetTag("s3.error", req.Error.Error())
	}
}
