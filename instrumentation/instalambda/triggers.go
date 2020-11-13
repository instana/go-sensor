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
	cloudWatchLogsEventType
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
		// CloudWatch Logs fields
		AWSLogs json.RawMessage `json:"awslogs"`
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
	case len(v.AWSLogs) != 0:
		return cloudWatchLogsEventType
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

func extractCloudWatchLogsTriggerTags(evt events.CloudwatchLogsEvent) opentracing.Tags {
	logs, err := evt.AWSLogs.Parse()
	if err != nil {
		return opentracing.Tags{
			"lambda.trigger":                "aws:cloudwatch.logs",
			"cloudwatch.logs.decodingError": err,
		}
	}

	var events []string
	for _, event := range logs.LogEvents {
		events = append(events, event.Message)
	}

	return opentracing.Tags{
		"lambda.trigger":         "aws:cloudwatch.logs",
		"cloudwatch.logs.group":  logs.LogGroup,
		"cloudwatch.logs.stream": logs.LogStream,
		"cloudwatch.logs.events": events,
	}
}
