// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"context"
	"errors"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var errUnknownS3Method = errors.New("s3 method not instrumented")

func injectAWSContextWithS3Span(tr instana.TracerLogger, ctx context.Context, params interface{}) context.Context {
	tags, err := extractS3Tags(params)
	if err != nil {
		if errors.Is(err, errUnknownS3Method) {
			tr.Logger().Error("failed to identify the s3 method: ", err.Error())
			return ctx
		}
	}

	// By design, we will abort the s3 span creation if a parent span is not identified.
	parent, ok := instana.SpanFromContext(ctx)
	if !ok {
		tr.Logger().Error("failed to retrieve the parent span. Aborting s3 child span creation.")
		return ctx
	}

	sp := tr.Tracer().StartSpan("s3",
		ext.SpanKindRPCClient,
		opentracing.ChildOf(parent.Context()),
		opentracing.Tags{
			s3Region: awsmiddleware.GetRegion(ctx),
		},
		tags,
	)

	return instana.ContextWithSpan(ctx, sp)
}

func finishS3Span(tr instana.TracerLogger, ctx context.Context, err error) {
	sp, ok := instana.SpanFromContext(ctx)
	if !ok {
		tr.Logger().Error("failed to retrieve the s3 child span from context.")
		return
	}

	defer sp.Finish()

	if err != nil {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(s3Error, err.Error())
	}
}

func extractS3Tags(params interface{}) (opentracing.Tags, error) {
	switch params := params.(type) {
	case *s3.CreateBucketInput:
		return opentracing.Tags{
			s3Op:     "createBucket",
			s3Bucket: stringDeRef(params.Bucket),
		}, nil
	case *s3.DeleteBucketInput:
		return opentracing.Tags{
			s3Op:     "deleteBucket",
			s3Bucket: stringDeRef(params.Bucket),
		}, nil
	case *s3.DeleteObjectInput:
		return opentracing.Tags{
			s3Op:     "delete",
			s3Bucket: stringDeRef(params.Bucket),
			s3Key:    stringDeRef(params.Key),
		}, nil
	case *s3.DeleteObjectsInput:
		return opentracing.Tags{
			s3Op:     "delete",
			s3Bucket: stringDeRef(params.Bucket),
		}, nil
	case *s3.GetObjectInput:
		return opentracing.Tags{
			s3Op:     "get",
			s3Bucket: stringDeRef(params.Bucket),
			s3Key:    stringDeRef(params.Key),
		}, nil
	case *s3.HeadObjectInput:
		return opentracing.Tags{
			s3Op:     "metadata",
			s3Bucket: stringDeRef(params.Bucket),
			s3Key:    stringDeRef(params.Key),
		}, nil
	case *s3.ListObjectsInput:
		return opentracing.Tags{
			s3Op:     "list",
			s3Bucket: stringDeRef(params.Bucket),
		}, nil
	case *s3.ListObjectsV2Input:
		return opentracing.Tags{
			s3Op:     "list",
			s3Bucket: stringDeRef(params.Bucket),
		}, nil
	case *s3.PutObjectInput:
		return opentracing.Tags{
			s3Op:     "put",
			s3Bucket: stringDeRef(params.Bucket),
			s3Key:    stringDeRef(params.Key),
		}, nil
	default:
		return nil, errUnknownS3Method
	}
}
