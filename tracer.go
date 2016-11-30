package instana

import (
	"time"

	bt "github.com/opentracing/basictracer-go"
	ot "github.com/opentracing/opentracing-go"
)

const (
	MAX_LOGS_PER_SPAN = 2
)

type tracerS struct {
	options        bt.Options
	textPropagator *textMapPropagator
}

func (r *tracerS) Inject(sc ot.SpanContext, format interface{}, carrier interface{}) error {
	switch format {
	case ot.TextMap, ot.HTTPHeaders:
		return r.textPropagator.inject(sc, carrier)
	}

	return ot.ErrUnsupportedFormat
}

func (r *tracerS) Extract(format interface{}, carrier interface{}) (ot.SpanContext, error) {
	switch format {
	case ot.TextMap, ot.HTTPHeaders:
		return r.textPropagator.extract(carrier)
	}

	return nil, ot.ErrUnsupportedFormat
}

func (r *tracerS) StartSpan(operationName string, opts ...ot.StartSpanOption) ot.Span {
	sso := ot.StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}

	return r.StartSpanWithOptions(operationName, sso)
}

func (r *tracerS) StartSpanWithOptions(operationName string, opts ot.StartSpanOptions) ot.Span {
	startTime := opts.StartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}

	tags := opts.Tags
	span := &spanS{}
Loop:
	for _, ref := range opts.References {
		switch ref.Type {
		case ot.ChildOfRef, ot.FollowsFromRef:
			refCtx := ref.ReferencedContext.(bt.SpanContext)
			span.raw.Context.TraceID = refCtx.TraceID
			span.raw.Context.SpanID = randomID()
			span.raw.Context.Sampled = refCtx.Sampled
			span.raw.ParentSpanID = refCtx.SpanID
			if l := len(refCtx.Baggage); l > 0 {
				span.raw.Context.Baggage = make(map[string]string, l)
				for k, v := range refCtx.Baggage {
					span.raw.Context.Baggage[k] = v
				}
			}

			break Loop
		}
	}

	if span.raw.Context.TraceID == 0 {
		span.raw.Context.SpanID = randomID()
		span.raw.Context.TraceID = span.raw.Context.SpanID
		span.raw.Context.Sampled = r.options.ShouldSample(span.raw.Context.TraceID)
	}

	return r.startSpanInternal(span, operationName, startTime, tags)
}

func (r *tracerS) startSpanInternal(span *spanS, operationName string, startTime time.Time, tags ot.Tags) ot.Span {
	span.tracer = r
	span.raw.Operation = operationName
	span.raw.Start = startTime
	span.raw.Duration = -1
	span.raw.Tags = tags

	return span
}

func shouldSample(traceId uint64) bool {
	return false
}

func NewTracer() ot.Tracer {
	return NewTracerWithOptions(&Options{})
}

func NewTracerWithOptions(options *Options) ot.Tracer {
	return NewTracerWithEverything(options, NewRecorder())
}

func NewTracerWithEverything(options *Options, recorder bt.SpanRecorder) ot.Tracer {
	InitSensor(options)
	ret := &tracerS{options: bt.Options{
		Recorder:       recorder,
		ShouldSample:   shouldSample,
		MaxLogsPerSpan: MAX_LOGS_PER_SPAN}}
	ret.textPropagator = &textMapPropagator{ret}

	return ret
}
