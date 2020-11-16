package instalambda

import (
	"context"
	"encoding/json"
	"net/http"
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
)

type wrappedHandler struct {
	lambda.Handler

	sensor *instana.Sensor
}

func TraceHandlerFunc(handlerFunc interface{}, sensor *instana.Sensor) *wrappedHandler {
	return WrapHandler(lambda.NewHandler(handlerFunc), sensor)
}

func WrapHandler(h lambda.Handler, sensor *instana.Sensor) *wrappedHandler {
	return &wrappedHandler{h, sensor}
}

func (h *wrappedHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return h.Handler.Invoke(ctx, payload)
	}

	opts := append([]opentracing.StartSpanOption{opentracing.Tags{
		"lambda.arn":     lc.InvokedFunctionArn + ":" + lambdacontext.FunctionVersion,
		"lambda.name":    lambdacontext.FunctionName,
		"lambda.version": lambdacontext.FunctionVersion,
	}}, h.triggerEventSpanOptions(payload)...)
	sp := h.sensor.Tracer().StartSpan("aws.lambda.entry", opts...)

	resp, err := h.Handler.Invoke(instana.ContextWithSpan(ctx, sp), payload)
	if err != nil {
		sp.LogFields(otlog.Error(err))
	}

	sp.Finish()

	// ensure that all collected data has been sent before the invokation is finished
	if tr, ok := h.sensor.Tracer().(instana.Tracer); ok {
		var i int
		for {
			if err := tr.Flush(context.Background()); err != nil {
				if err == instana.ErrAgentNotReady && i < awsLambdaFlushMaxRetries {
					i++
					time.Sleep(awsLambdaFlushRetryPeriod)
					continue
				}

				h.sensor.Logger().Error("failed to send traces:", err)
			}

			break
		}
	}

	return resp, err
}

func (h *wrappedHandler) triggerEventSpanOptions(payload []byte) []opentracing.StartSpanOption {
	switch detectTriggerEventType(payload) {
	case apiGatewayEventType:
		var v events.APIGatewayProxyRequest
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal API Gateway event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		opts := []opentracing.StartSpanOption{extractAPIGatewayTriggerTags(v)}
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

		opts := []opentracing.StartSpanOption{extractAPIGatewayV2TriggerTags(v)}
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

		opts := []opentracing.StartSpanOption{extractALBTriggerTags(v)}
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

		return []opentracing.StartSpanOption{extractCloudWatchTriggerTags(v)}
	case cloudWatchLogsEventType:
		var v events.CloudwatchLogsEvent
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal CloudWatch Logs event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		return []opentracing.StartSpanOption{extractCloudWatchLogsTriggerTags(v)}
	case s3EventType:
		var v events.S3Event
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal S3 event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		return []opentracing.StartSpanOption{extractS3TriggerTags(v)}
	case sqsEventType:
		var v events.SQSEvent
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal SQS event payload: ", err)
			return []opentracing.StartSpanOption{opentracing.Tags{}}
		}

		return []opentracing.StartSpanOption{extractSQSTriggerTags(v)}
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
