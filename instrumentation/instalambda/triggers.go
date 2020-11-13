package instalambda

import (
	"encoding/json"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/opentracing/opentracing-go"
)

type triggerEventType uint8

const (
	unknownEventType triggerEventType = iota
	apiGatewayEventType
	albEventType
	cloudWatchEventType
)

func detectTriggerEventType(payload []byte) triggerEventType {
	var v struct {
		// API Gateway fields
		Resource   string `json:"resource"`
		Path       string `json:"path"`
		HTTPMethod string `json:"httpMethod"`
		// ALB fields
		RequestContext struct {
			ELB json.RawMessage `json:"elb"`
		} `json:"requestContext"`
		// CloudWatch fields
		Source     string `json:"source"`
		DetailType string `json:"detail-type"`
	}

	if err := json.Unmarshal(payload, &v); err != nil {
		return unknownEventType
	}

	switch {
	case v.Resource != "" && v.Path != "" && v.HTTPMethod != "" && v.RequestContext.ELB == nil:
		return apiGatewayEventType
	case v.RequestContext.ELB != nil:
		return albEventType
	case v.Source == "aws.events" && v.DetailType == "Scheduled Event":
		return cloudWatchEventType
	default:
		return unknownEventType
	}
}

func extractAPIGatewayTriggerTags(evt events.APIGatewayProxyRequest) opentracing.Tags {
	params := url.Values{}

	for k, v := range evt.QueryStringParameters {
		params.Set(k, v)
	}

	for k, vv := range evt.MultiValueQueryStringParameters {
		for _, v := range vv {
			params.Add(k, v)
		}
	}

	return opentracing.Tags{
		"lambda.trigger": "aws:api.gateway",
		"http.method":    evt.HTTPMethod,
		"http.url":       evt.Path,
		"http.path_tpl":  evt.Resource,
		"http.params":    params.Encode(),
	}
}

func extractALBTriggerTags(evt events.ALBTargetGroupRequest) opentracing.Tags {
	params := url.Values{}

	for k, v := range evt.QueryStringParameters {
		params.Set(k, v)
	}

	for k, vv := range evt.MultiValueQueryStringParameters {
		for _, v := range vv {
			params.Add(k, v)
		}
	}

	return opentracing.Tags{
		"lambda.trigger": "aws:application.load.balancer",
		"http.method":    evt.HTTPMethod,
		"http.url":       evt.Path,
		"http.params":    params.Encode(),
	}
}

func extractCloudWatchTriggerTags(evt events.CloudWatchEvent) opentracing.Tags {
	return opentracing.Tags{
		"lambda.trigger":              "aws:cloudwatch.events",
		"cloudwatch.events.id":        evt.ID,
		"cloudwatch.events.resources": evt.Resources,
	}
}
