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

// StartS3Span initiates a new span from an AWS S3 request and injects it into the
// request.Request context
func StartS3Span(req *request.Request, sensor *instana.Sensor) {
	tags, err := extractS3Tags(req)
	if err != nil {
		if err == errMethodNotInstrumented {
			return
		}

		sensor.Logger().Warn("failed to extract S3 tags: ", err)
	}

	parent, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}

	sp := sensor.Tracer().StartSpan("s3",
		ext.SpanKindRPCClient,
		opentracing.ChildOf(parent.Context()),
		opentracing.Tags{
			s3Region: req.ClientInfo.SigningRegion,
		},
		tags,
	)

	req.SetContext(instana.ContextWithSpan(req.Context(), sp))
}

// FinalizeS3Span retrieves tags from completed request.Request and adds them
// to the span
func FinalizeS3Span(req *request.Request) {
	sp, ok := instana.SpanFromContext(req.Context())
	if !ok {
		return
	}
	defer sp.Finish()

	if req.Error != nil {
		sp.LogFields(otlog.Error(req.Error))
		sp.SetTag(s3Error, req.Error.Error())
	}
}
