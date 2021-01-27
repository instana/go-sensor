// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/opentracing/opentracing-go"
)

func extractSNSTags(req *request.Request) (opentracing.Tags, error) {
	switch params := req.Params.(type) {
	case *sns.PublishInput:
		return opentracing.Tags{
			"sns.topic":   aws.StringValue(params.TopicArn),
			"sns.target":  aws.StringValue(params.TargetArn),
			"sns.phone":   aws.StringValue(params.PhoneNumber),
			"sns.subject": aws.StringValue(params.Subject),
		}, nil
	default:
		return nil, errMethodNotInstrumented
	}
}
