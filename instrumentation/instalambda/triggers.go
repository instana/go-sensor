package instalambda

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
)

type triggerEventType uint8

const (
	unknownEventType triggerEventType = iota
	apiGatewayEventType
	apiGatewayV2EventType
	albEventType
	cloudWatchEventType
	cloudWatchLogsEventType
	s3EventType
	sqsEventType
)

func detectTriggerEventType(payload []byte) triggerEventType {
	var v struct {
		// API Gateway fields
		Resource   string `json:"resource"`
		Path       string `json:"path"`
		HTTPMethod string `json:"httpMethod"`
		// CloudWatch fields
		Source     string `json:"source"`
		DetailType string `json:"detail-type"`
		// CloudWatch Logs fields
		AWSLogs json.RawMessage `json:"awslogs"`
		// S3 and SQS fields
		Records []struct {
			Source string `json:"eventSource"`
		}
		// Version is common for multiple event types
		Version string `json:"version"`
		// RequestContext is common for multiple event types
		RequestContext struct {
			// ALB fields
			ELB json.RawMessage `json:"elb"`
			// API Gateway v2.0 fields
			ApiID string          `json:"apiId"`
			Stage string          `json:"stage"`
			HTTP  json.RawMessage `json:"http"`
		} `json:"requestContext"`
	}

	if err := json.Unmarshal(payload, &v); err != nil {
		return unknownEventType
	}

	switch {
	case v.Resource != "" && v.Path != "" && v.HTTPMethod != "" && v.RequestContext.ELB == nil:
		return apiGatewayEventType
	case v.Version == "2.0" && v.RequestContext.ApiID != "" && v.RequestContext.Stage != "" && len(v.RequestContext.HTTP) > 0:
		return apiGatewayV2EventType
	case v.RequestContext.ELB != nil:
		return albEventType
	case v.Source == "aws.events" && v.DetailType == "Scheduled Event":
		return cloudWatchEventType
	case len(v.AWSLogs) != 0:
		return cloudWatchLogsEventType
	case len(v.Records) > 0 && v.Records[0].Source == "aws:s3":
		return s3EventType
	case len(v.Records) > 0 && v.Records[0].Source == "aws:sqs":
		return sqsEventType
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

func extractAPIGatewayV2TriggerTags(evt events.APIGatewayV2HTTPRequest) opentracing.Tags {
	params := url.Values{}

	for k, v := range evt.QueryStringParameters {
		params.Set(k, v)
	}

	routeKeyPath := evt.RouteKey
	// Strip any trailing HTTP request method
	if i := strings.Index(routeKeyPath, " "); i >= 0 {
		routeKeyPath = evt.RouteKey[i+1:]
	}

	return opentracing.Tags{
		"lambda.trigger": "aws:api.gateway",
		"http.method":    evt.RequestContext.HTTP.Method,
		"http.url":       evt.RequestContext.HTTP.Path,
		"http.path_tpl":  routeKeyPath,
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

func extractS3TriggerTags(evt events.S3Event) opentracing.Tags {
	var events []instana.AWSS3EventTags
	for _, rec := range evt.Records {
		events = append(events, instana.AWSS3EventTags{
			Name:   rec.EventName,
			Bucket: rec.S3.Bucket.Name,
			Object: rec.S3.Object.Key,
		})
	}

	return opentracing.Tags{
		"lambda.trigger": "aws:s3",
		"s3.events":      events,
	}
}

func extractSQSTriggerTags(evt events.SQSEvent) opentracing.Tags {
	var msgs []instana.AWSSQSMessageTags
	for _, rec := range evt.Records {
		msgs = append(msgs, instana.AWSSQSMessageTags{
			Queue: rec.EventSourceARN,
		})
	}

	return opentracing.Tags{
		"lambda.trigger": "aws:sqs",
		"sqs.messages":   msgs,
	}
}
