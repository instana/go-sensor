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
	}

	if err := json.Unmarshal(payload, &v); err != nil {
		return unknownEventType
	}

	switch {
	case v.Resource != "" && v.Path != "" && v.HTTPMethod != "" && v.RequestContext.ELB == nil:
		return apiGatewayEventType
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
