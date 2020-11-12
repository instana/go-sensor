package instalambda

import (
	"context"
	"time"

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
	}, extractTriggerEventTags(payload))

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
