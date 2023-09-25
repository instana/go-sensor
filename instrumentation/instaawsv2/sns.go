// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var errUnknownSNSMethod = errors.New("sns method not instrumented")

type AWSSNSOperations struct{}

var _ AWSOperations = (*AWSSNSOperations)(nil)

func (o AWSSNSOperations) injectContextWithSpan(tr instana.TracerLogger, ctx context.Context, params interface{}) context.Context {
	tags, err := o.extractTags(params)
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
		ot.ChildOf(parent.Context()),
		tags,
	)

	if err = o.injectSpanToCarrier(params, sp); err != nil {
		tr.Logger().Error("failed to inject span context to the sns carrier: ", err.Error())
	}

	return instana.ContextWithSpan(ctx, sp)
}

func (o AWSSNSOperations) finishSpan(tr instana.TracerLogger, ctx context.Context, err error) {
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

func (o AWSSNSOperations) injectSpanToCarrier(params interface{}, sp ot.Span) error {
	var ip *sns.PublishInput
	var ok bool

	if ip, ok = params.(*sns.PublishInput); !ok {
		return errors.New("received param is not of type sns.PublishInput")
	}

	if ip.MessageAttributes == nil {
		ip.MessageAttributes = make(map[string]types.MessageAttributeValue)
	}

	err := sp.Tracer().Inject(sp.Context(), ot.TextMap, snsMessageAttributesCarrier(ip.MessageAttributes))
	if err != nil {
		return err
	}

	return nil
}

func (o AWSSNSOperations) extractTags(params interface{}) (ot.Tags, error) {
	switch params := params.(type) {
	case *sns.PublishInput:
		return ot.Tags{
			snsTopic:   stringDeRef(params.TopicArn),
			snsTarget:  stringDeRef(params.TargetArn),
			snsPhone:   stringDeRef(params.PhoneNumber),
			snsSubject: stringDeRef(params.Subject),
		}, nil
	default:
		return nil, errUnknownSNSMethod
	}
}
