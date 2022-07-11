// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

// Package instalambda provides Instana tracing instrumentation for
// AWS Lambda functions
package instalambda

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

const (
	awsLambdaFlushMaxRetries  = 5
	awsLambdaFlushRetryPeriod = 50 * time.Millisecond
	awsLambdaTimeoutThreshold = 100 * time.Millisecond
)

var errHandlerTimedOut = errors.New("handler has timed out")

type wrappedHandler struct {
	lambda.Handler

	sensor      *instana.Sensor
	onColdStart sync.Once
}

// NewHandler creates a new instrumented handler that can be used with `lambda.StartHandler()` from a handler function
func NewHandler(handlerFunc interface{}, sensor *instana.Sensor) *wrappedHandler {
	return WrapHandler(lambda.NewHandler(handlerFunc), sensor)
}

// WrapHandler instruments a lambda.Handler to trace the invokations with Instana
func WrapHandler(h lambda.Handler, sensor *instana.Sensor) *wrappedHandler {
	return &wrappedHandler{
		Handler: h,
		sensor:  sensor,
	}
}

// Invoke is a handler function for a wrapped handler
func (h *wrappedHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return h.Handler.Invoke(ctx, payload)
	}

	opts := append([]opentracing.StartSpanOption{opentracing.Tags{
		lambdaArn:     lc.InvokedFunctionArn + ":" + lambdacontext.FunctionVersion,
		lambdaName:    lambdacontext.FunctionName,
		lambdaVersion: lambdacontext.FunctionVersion,
	}}, h.triggerEventSpanOptions(payload, lc.ClientContext)...)

	sp := h.sensor.Tracer().StartSpan("aws.lambda.entry", opts...)
	h.onColdStart.Do(func() {
		sp.SetTag(lambdaColdStart, true)
	})

	// Here we create a separate context.Context to finalize and send the span. This context
	// supposed to be canceled once the wrapped handler is done.
	traceCtx, cancelTraceCtx := context.WithCancel(ctx)

	// In case there is a deadline defined for this invokation, we should finish the span just before
	// the function times out to send the span data.
	originalDeadline, deadlineDefined := ctx.Deadline()
	if deadlineDefined {
		traceCtx, cancelTraceCtx = context.WithDeadline(ctx, originalDeadline.Add(-awsLambdaTimeoutThreshold))
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Await for the trace context to become either canceled or timed out and finalize the span
	go func() {
		defer wg.Done()

		<-traceCtx.Done()

		if traceCtx.Err() == context.DeadlineExceeded {
			remainingTime := time.Until(originalDeadline).Truncate(time.Millisecond)
			h.sensor.Logger().Debug("heuristical timeout detection was triggered with ", remainingTime, " left")

			sp.SetTag(lambdaMsLeft, int64(remainingTime)/1e6) // cast time.Duration to int64 for compatibility with older Go versions
			sp.LogFields(otlog.Error(errHandlerTimedOut))
		}

		sp.Finish()
		h.flushAgent(awsLambdaFlushRetryPeriod, awsLambdaFlushMaxRetries)
	}()

	resp, err := h.Handler.Invoke(instana.ContextWithSpan(ctx, sp), payload)
	if err != nil {
		sp.LogFields(otlog.Error(err))
	}

	cancelTraceCtx()

	// ensure that span has been finished and sent to the agent before quit
	wg.Wait()

	return resp, err
}

func (h *wrappedHandler) flushAgent(retryPeriod time.Duration, maxRetries int) {
	h.sensor.Logger().Debug("flushing trace data")

	if tr, ok := h.sensor.Tracer().(instana.Tracer); ok {
		var i int
		for {
			if err := tr.Flush(context.Background()); err != nil {
				if err == instana.ErrAgentNotReady && i < maxRetries {
					i++
					time.Sleep(retryPeriod)
					continue
				}

				h.sensor.Logger().Error("failed to send traces:", err)
			}

			h.sensor.Logger().Debug("data sent")
			break
		}
	}
}

func (h *wrappedHandler) triggerEventSpanOptions(payload []byte, lcc lambdacontext.ClientContext) []opentracing.StartSpanOption {
	switch detectTriggerEventType(payload) {
	case apiGatewayEventType:
		var v events.APIGatewayProxyRequest
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal API Gateway event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		opts := []opentracing.StartSpanOption{h.extractAPIGatewayTriggerTags(v)}
		if parentCtx, ok := h.extractParentContext(v.Headers); ok {
			opts = append(opts, opentracing.ChildOf(parentCtx))
		}

		return opts
	case apiGatewayV2EventType:
		var v events.APIGatewayV2HTTPRequest
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal API Gateway v2.0 event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		opts := []opentracing.StartSpanOption{h.extractAPIGatewayV2TriggerTags(v)}
		if parentCtx, ok := h.extractParentContext(v.Headers); ok {
			opts = append(opts, opentracing.ChildOf(parentCtx))
		}

		return opts
	case albEventType:
		var v events.ALBTargetGroupRequest
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal ALB event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		opts := []opentracing.StartSpanOption{h.extractALBTriggerTags(v)}
		if parentCtx, ok := h.extractParentContext(v.Headers); ok {
			opts = append(opts, opentracing.ChildOf(parentCtx))
		}

		return opts
	case cloudWatchEventType:
		var v events.CloudWatchEvent
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal CloudWatch event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		return []opentracing.StartSpanOption{h.extractCloudWatchTriggerTags(v)}
	case cloudWatchLogsEventType:
		var v events.CloudwatchLogsEvent
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal CloudWatch Logs event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		return []opentracing.StartSpanOption{h.extractCloudWatchLogsTriggerTags(v)}
	case s3EventType:
		var v events.S3Event
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal S3 event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		return []opentracing.StartSpanOption{h.extractS3TriggerTags(v)}
	case sqsEventType:
		var v events.SQSEvent
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal SQS event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		return []opentracing.StartSpanOption{h.extractSQSTriggerTags(v)}
	case invokeRequestType:

		tags := opentracing.Tags{
			lambdaTrigger: "aws:lambda.invoke",
		}

		opts := []opentracing.StartSpanOption{tags}
		if parentCtx, ok := h.extractParentContext(lcc.Custom); ok {
			opts = append(opts, opentracing.ChildOf(parentCtx))
		}
		return opts

	default:
		h.sensor.Logger().Info("unsupported AWS Lambda trigger event type, the entry span will include generic tags only")
		return []opentracing.StartSpanOption{opentracing.Tags{}}
	}
}

func (h *wrappedHandler) extractParentContext(headers map[string]string) (opentracing.SpanContext, bool) {
	hdrs := http.Header{}
	for k, v := range headers {
		hdrs.Set(k, v)
	}

	switch parentCtx, err := h.sensor.Tracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdrs)); err {
	case nil:
		return parentCtx, true
	case opentracing.ErrSpanContextNotFound:
		h.sensor.Logger().Debug("lambda invoke event did not provide trace context")
	case opentracing.ErrUnsupportedFormat:
		h.sensor.Logger().Info("lambda invoke event provided trace context in unsupported format")
	default:
		h.sensor.Logger().Warn("failed to extract span context from the lambda invoke event:", err)
	}

	return nil, false
}

func (h *wrappedHandler) extractAPIGatewayTriggerTags(evt events.APIGatewayProxyRequest) opentracing.Tags {
	tags := opentracing.Tags{
		lambdaTrigger: "aws:api.gateway",
		httpMethod:    evt.HTTPMethod,
		httpUrl:       evt.Path,
		httpPathTpl:   evt.Resource,
		httpParams:    h.sanitizeHTTPParams(evt.QueryStringParameters, evt.MultiValueQueryStringParameters).Encode(),
	}

	if headers := h.collectHTTPHeaders(evt.Headers, evt.MultiValueHeaders); len(headers) > 0 {
		tags[httpHeader] = headers
	}

	return tags
}

func (h *wrappedHandler) extractAPIGatewayV2TriggerTags(evt events.APIGatewayV2HTTPRequest) opentracing.Tags {
	routeKeyPath := evt.RouteKey
	// Strip any trailing HTTP request method
	if i := strings.Index(routeKeyPath, " "); i >= 0 {
		routeKeyPath = evt.RouteKey[i+1:]
	}

	tags := opentracing.Tags{
		lambdaTrigger: "aws:api.gateway",
		httpMethod:    evt.RequestContext.HTTP.Method,
		httpUrl:       evt.RequestContext.HTTP.Path,
		httpPathTpl:   routeKeyPath,
		httpParams:    h.sanitizeHTTPParams(evt.QueryStringParameters, nil).Encode(),
	}

	if headers := h.collectHTTPHeaders(evt.Headers, nil); len(headers) > 0 {
		tags[httpHeader] = headers
	}

	return tags
}

func (h *wrappedHandler) extractALBTriggerTags(evt events.ALBTargetGroupRequest) opentracing.Tags {
	tags := opentracing.Tags{
		lambdaTrigger: "aws:application.load.balancer",
		httpMethod:    evt.HTTPMethod,
		httpUrl:       evt.Path,
		httpParams:    h.sanitizeHTTPParams(evt.QueryStringParameters, evt.MultiValueQueryStringParameters).Encode(),
	}

	if headers := h.collectHTTPHeaders(evt.Headers, evt.MultiValueHeaders); len(headers) > 0 {
		tags[httpHeader] = headers
	}

	return tags
}

func (h *wrappedHandler) extractCloudWatchTriggerTags(evt events.CloudWatchEvent) opentracing.Tags {
	return opentracing.Tags{
		lambdaTrigger:             "aws:cloudwatch.events",
		cloudwatchEventsId:        evt.ID,
		cloudwatchEventsResources: evt.Resources,
	}
}

func (h *wrappedHandler) extractCloudWatchLogsTriggerTags(evt events.CloudwatchLogsEvent) opentracing.Tags {
	logs, err := evt.AWSLogs.Parse()
	if err != nil {
		return opentracing.Tags{
			lambdaTrigger:               "aws:cloudwatch.logs",
			cloudwatchLogsDecodingError: err,
		}
	}

	var e []string
	for _, event := range logs.LogEvents {
		e = append(e, event.Message)
	}

	return opentracing.Tags{
		lambdaTrigger:        "aws:cloudwatch.logs",
		cloudwatchLogsGroup:  logs.LogGroup,
		cloudwatchLogsStream: logs.LogStream,
		cloudwatchLogsEvents: e,
	}
}

func (h *wrappedHandler) extractS3TriggerTags(evt events.S3Event) opentracing.Tags {
	var e []instana.AWSS3EventTags
	for _, rec := range evt.Records {
		e = append(e, instana.AWSS3EventTags{
			Name:   rec.EventName,
			Bucket: rec.S3.Bucket.Name,
			Object: rec.S3.Object.Key,
		})
	}

	return opentracing.Tags{
		lambdaTrigger: "aws:s3",
		s3Events:      e,
	}
}

func (h *wrappedHandler) extractSQSTriggerTags(evt events.SQSEvent) opentracing.Tags {
	var msgs []instana.AWSSQSMessageTags
	for _, rec := range evt.Records {
		msgs = append(msgs, instana.AWSSQSMessageTags{
			Queue: rec.EventSourceARN,
		})
	}

	return opentracing.Tags{
		lambdaTrigger: "aws:sqs",
		sqsMessages:   msgs,
	}
}

func (h *wrappedHandler) sanitizeHTTPParams(
	queryStringParams map[string]string,
	multiValueQueryStringParams map[string][]string,
) url.Values {
	secretMatcher := instana.DefaultSecretsMatcher()
	if tr, ok := h.sensor.Tracer().(instana.Tracer); ok {
		secretMatcher = tr.Options().Secrets
	}

	params := url.Values{}

	for k, v := range queryStringParams {
		if secretMatcher.Match(k) {
			v = "<redacted>"
		}
		params.Set(k, v)
	}

	for k, vv := range multiValueQueryStringParams {
		isSecret := secretMatcher.Match(k)
		for _, v := range vv {
			if isSecret {
				v = "<redacted>"
			}
			params.Add(k, v)
		}
	}

	return params
}

func (h *wrappedHandler) collectHTTPHeaders(headers map[string]string, multiValueHeaders map[string][]string) map[string]string {
	var collectableHeaders []string
	if tr, ok := h.sensor.Tracer().(instana.Tracer); ok {
		collectableHeaders = tr.Options().CollectableHTTPHeaders
	}

	if len(collectableHeaders) == 0 {
		return nil
	}

	// Normalize header names first by bringing them to the canonical MIME format to avoid missing headers because of mismatching case
	hdrs := http.Header{}
	for k, v := range headers {
		hdrs.Set(k, v)
	}

	for k, vv := range multiValueHeaders {
		for _, v := range vv {
			hdrs.Add(k, v)
		}
	}

	collected := make(map[string]string)
	for _, k := range collectableHeaders {
		if v := hdrs.Get(k); v != "" {
			collected[k] = v
		}
	}

	return collected
}
