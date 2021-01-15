// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/opentracing/opentracing-go"
)

func extractS3Tags(req *request.Request) (opentracing.Tags, error) {
	switch params := req.Params.(type) {
	case *s3.CreateBucketInput:
		return opentracing.Tags{
			"s3.op":     "createBucket",
			"s3.bucket": aws.StringValue(params.Bucket),
		}, nil
	case *s3.DeleteBucketInput:
		return opentracing.Tags{
			"s3.op":     "deleteBucket",
			"s3.bucket": aws.StringValue(params.Bucket),
		}, nil
	case *s3.DeleteObjectInput:
		return opentracing.Tags{
			"s3.op":     "delete",
			"s3.bucket": aws.StringValue(params.Bucket),
			"s3.key":    aws.StringValue(params.Key),
		}, nil
	case *s3.DeleteObjectsInput:
		return opentracing.Tags{
			"s3.op":     "delete",
			"s3.bucket": aws.StringValue(params.Bucket),
		}, nil
	case *s3.GetObjectInput:
		return opentracing.Tags{
			"s3.op":     "get",
			"s3.bucket": aws.StringValue(params.Bucket),
			"s3.key":    aws.StringValue(params.Key),
		}, nil
	case *s3.HeadObjectInput:
		return opentracing.Tags{
			"s3.op":     "metadata",
			"s3.bucket": aws.StringValue(params.Bucket),
			"s3.key":    aws.StringValue(params.Key),
		}, nil
	case *s3.ListObjectsInput:
		return opentracing.Tags{
			"s3.op":     "list",
			"s3.bucket": aws.StringValue(params.Bucket),
		}, nil
	case *s3.ListObjectsV2Input:
		return opentracing.Tags{
			"s3.op":     "list",
			"s3.bucket": aws.StringValue(params.Bucket),
		}, nil
	case *s3.PutObjectInput:
		return opentracing.Tags{
			"s3.op":     "put",
			"s3.bucket": aws.StringValue(params.Bucket),
			"s3.key":    aws.StringValue(params.Key),
		}, nil
	default:
		return nil, errMethodNotInstrumented
	}
}
