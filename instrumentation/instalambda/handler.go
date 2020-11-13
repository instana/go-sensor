package instalambda

import (
	"context"
	"encoding/json"
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

	sp := h.sensor.Tracer().StartSpan("aws.lambda.entry", opentracing.Tags{
		"lambda.arn":     lc.InvokedFunctionArn + ":" + lambdacontext.FunctionVersion,
		"lambda.name":    lambdacontext.FunctionName,
		"lambda.version": lambdacontext.FunctionVersion,
	}, h.extractTriggerEventTags(payload))

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

func (h *wrappedHandler) extractTriggerEventTags(payload []byte) opentracing.Tags {
	switch detectTriggerEventType(payload) {
	case apiGatewayEventType:
		var v events.APIGatewayProxyRequest
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal API Gateway event payload: ", err)
			return opentracing.Tags{}
		}

		return extractAPIGatewayTriggerTags(v)
	case albEventType:
		var v events.ALBTargetGroupRequest
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal ALB event payload: ", err)
			return opentracing.Tags{}
		}

		return extractALBTriggerTags(v)
	case cloudWatchEventType:
		var v events.CloudWatchEvent
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal CloudWatch event payload: ", err)
			return opentracing.Tags{}
		}

		return extractCloudWatchTriggerTags(v)
	case cloudWatchLogsEventType:
		var v events.CloudwatchLogsEvent
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal CloudWatch Logs event payload: ", err)
			return opentracing.Tags{}
		}

		return extractCloudWatchLogsTriggerTags(v)
	case s3EventType:
		var v events.S3Event
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal S3 event payload: ", err)
			return opentracing.Tags{}
		}

		return extractS3TriggerTags(v)
	case sqsEventType:
		var v events.SQSEvent
		if err := json.Unmarshal(payload, &v); err != nil {
			h.sensor.Logger().Warn("failed to unmarshal SQS event payload: ", err)
			return opentracing.Tags{}
		}

		return extractSQSTriggerTags(v)
	default:
		h.sensor.Logger().Info("unsupported AWS Lambda trigger event type, the entry span will include generic tags only")
		return opentracing.Tags{}
	}
}
