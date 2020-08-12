package instana

import (
	"time"

	ot "github.com/opentracing/opentracing-go"
)

const (
	// MaxLogsPerSpan The maximum number of logs allowed on a span.
	MaxLogsPerSpan = 2
)

type tracerS struct {
	recorder SpanRecorder
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
		return extractTraceContext(carrier)
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
			parent := ref.ReferencedContext.(SpanContext)
			corrData = parent.Correlation
			sc = NewSpanContext(parent)
			break
		}
	}

	if tag, ok := opts.Tags[suppressTracingTag]; ok {
		sc.Suppressed = tag.(bool)
		delete(opts.Tags, suppressTracingTag)
	}

	return &spanS{
		context:     sc,
		tracer:      r,
		Service:     sensor.options.Service,
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

// NewTracer initializes a new tracer with default options
func NewTracer() ot.Tracer {
	return NewTracerWithOptions(nil)
}

// NewTracerWithOptions initializes and configures a new tracer that collects and sends spans to the host agent
func NewTracerWithOptions(options *Options) ot.Tracer {
	return NewTracerWithEverything(options, NewRecorder())
}

// NewTracerWithEverything initializes and configures a new tracer
func NewTracerWithEverything(options *Options, recorder SpanRecorder) ot.Tracer {
	InitSensor(options)

	tracer := &tracerS{
		recorder: recorder,
	}

	return tracer
}
