// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instalambda

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

type triggerEventType uint8

const (
	// FieldT is the trace ID attribute key in a custom client context
	fieldT = "X-INSTANA-T"
	// FieldS is the span ID attribute key in a custom client context
	fieldS = "X-INSTANA-S"
	// FieldL is the trace level attribute key in a custom client context
	fieldL = "X-INSTANA-L"

	// traceParentHeader is the W3C trace parent header name as defined by https://www.w3.org/TR/trace-context/
	traceParentHeader = "TRACEPARENT"
	// TraceStateHeader is the W3C trace state header name as defined by https://www.w3.org/TR/trace-context/
	traceStateHeader = "TRACESTATE"

	unknownEventType triggerEventType = iota
	apiGatewayEventType
	apiGatewayV2EventType
	albEventType
	cloudWatchEventType
	cloudWatchLogsEventType
	s3EventType
	sqsEventType
	sdkInvokeRequestType
)

func detectTriggerEventType(payload []byte, lcc lambdacontext.ClientContext) triggerEventType {
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
	case areTracingHeadersInTheCustomContext(lcc):
		return sdkInvokeRequestType
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

func areTracingHeadersInTheCustomContext(lcc lambdacontext.ClientContext) bool {
	if lcc.Custom == nil {
		return false
	}

	normalizedCustomKeys := make(map[string]string, len(lcc.Custom))

	for k := range lcc.Custom {
		normalizedCustomKeys[strings.ToUpper(k)] = k
	}

	_, okS := normalizedCustomKeys[fieldS]
	_, okT := normalizedCustomKeys[fieldT]
	_, okL := normalizedCustomKeys[fieldL]

	_, okW3CTParent := normalizedCustomKeys[traceParentHeader]
	_, okW3CTState := normalizedCustomKeys[traceStateHeader]

	return (okS && okT && okL) || (okW3CTParent && okW3CTState)
}
