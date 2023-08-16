// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var errUnknownSNSMethod = errors.New("sns method not instrumented")

func injectAWSContextWithSNSSpan(tr instana.TracerLogger, ctx context.Context, params interface{}) context.Context {
	tags, err := extractSNSTags(params)
	if err != nil {
		if errors.Is(err, errUnknownSNSMethod) {
			tr.Logger().Error("failed to identify the sqs method: ", err.Error())
			return ctx
		}
	}

	// By design, we will abort the sns span creation if a parent span is not identified.
	parent, ok := instana.SpanFromContext(ctx)
	if !ok {
		tr.Logger().Error("failed to retrieve the parent span. Aborting sqs child span creation.")
		return ctx
	}

	sp := tr.Tracer().StartSpan("sns",
		ext.SpanKindRPCClient,
		opentracing.ChildOf(parent.Context()),
		tags,
	)

	injectSNSSpantoCarrier(params, sp)

	return instana.ContextWithSpan(ctx, sp)
}

func finishSNSSpan(tr instana.TracerLogger, ctx context.Context, err error) {
	sp, ok := instana.SpanFromContext(ctx)
	if !ok {
		tr.Logger().Error("failed to retrieve the sns child span from context.")
		return
	}
	defer sp.Finish()

	if err != nil {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(snsError, err.Error())
	}
}

func injectSNSSpantoCarrier(params interface{}, sp opentracing.Span) {
	var ip *sns.PublishInput
	var ok bool

	if ip, ok = params.(*sns.PublishInput); !ok {
		return
	}

	if ip.MessageAttributes == nil {
		ip.MessageAttributes = make(map[string]types.MessageAttributeValue)
	}

	sp.Tracer().Inject(sp.Context(), opentracing.TextMap, snsMessageAttributesCarrier(ip.MessageAttributes))
}

func extractSNSTags(params interface{}) (opentracing.Tags, error) {
	switch params := params.(type) {
	case *sns.PublishInput:
		return opentracing.Tags{
			snsTopic:   stringDeRef(params.TopicArn),
			snsTarget:  stringDeRef(params.TargetArn),
			snsPhone:   stringDeRef(params.PhoneNumber),
			snsSubject: stringDeRef(params.Subject),
		}, nil
	default:
		return nil, errUnknownSNSMethod
	}
}
