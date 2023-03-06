// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"context"
	"time"

	ot "github.com/opentracing/opentracing-go"
)

const (
	// MaxLogsPerSpan The maximum number of logs allowed on a span.
	MaxLogsPerSpan = 2
)

var _ ot.Tracer = (*tracerS)(nil)
var _ Tracer = (*tracerS)(nil)

type tracerS struct {
	recorder SpanRecorder
}

// NewTracer initializes a new tracer with default options
func NewTracer() *tracerS {
	return NewTracerWithOptions(nil)
}

// NewTracerWithOptions initializes and configures a new tracer that collects and sends spans to the agent
func NewTracerWithOptions(options *Options) *tracerS {
	return NewTracerWithEverything(options, nil)
}

// NewTracerWithEverything initializes and configures a new tracer. It uses instana.DefaultOptions() if nil
// is provided
func NewTracerWithEverything(options *Options, recorder SpanRecorder) *tracerS {
	InitSensor(options)

	if recorder == nil {
		recorder = NewRecorder()
	}

	tracer := &tracerS{
		recorder: recorder,
	}

	return tracer
}

func (r *tracerS) Inject(spanContext ot.SpanContext, format interface{}, carrier interface{}) error {
	switch format {
	case ot.TextMap, ot.HTTPHeaders:
		sc, ok := spanContext.(SpanContext)
		if !ok {
			return ot.ErrInvalidSpanContext
		}

		return injectTraceContext(sc, carrier)
	}

	return ot.ErrUnsupportedFormat
}

func (r *tracerS) Extract(format interface{}, carrier interface{}) (ot.SpanContext, error) {
	switch format {
	case ot.TextMap, ot.HTTPHeaders:
		sc, err := extractTraceContext(carrier)
		if err != nil {
			return nil, err
		}

		return sc, nil
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

	var corrData EUMCorrelationData

	sc := NewRootSpanContext()
	for _, ref := range opts.References {
		if ref.Type == ot.ChildOfRef || ref.Type == ot.FollowsFromRef {
			if parent, ok := ref.ReferencedContext.(SpanContext); ok {
				corrData = parent.Correlation
				sc = NewSpanContext(parent)
				break
			}
		}
	}

	if tag, ok := opts.Tags[suppressTracingTag]; ok {
		sc.Suppressed = tag.(bool)
		delete(opts.Tags, suppressTracingTag)
	}

	return &spanS{
		context:     sc,
		tracer:      r,
		Service:     sensor.serviceName,
		Operation:   operationName,
		Start:       startTime,
		Duration:    -1,
		Correlation: corrData,
		Tags:        opts.Tags,
	}
}

// Options returns current tracer options
func (r *tracerS) Options() TracerOptions {
	if sensor.options == nil {
		return DefaultTracerOptions()
	}

	return sensor.options.Tracer
}

// Flush forces sending any queued finished spans to the agent
func (r *tracerS) Flush(ctx context.Context) error {
	if err := r.recorder.Flush(ctx); err != nil {
		return err
	}

	return sensor.Agent().Flush(ctx)
}
