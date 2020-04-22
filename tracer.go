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
	options TracerOptions
}

func (r *tracerS) Inject(sc ot.SpanContext, format interface{}, carrier interface{}) error {
	switch format {
	case ot.TextMap, ot.HTTPHeaders:
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

	span := &spanS{
		tracer:    r,
		Service:   sensor.serviceName,
		Operation: operationName,
		Start:     startTime,
		Duration:  -1,
		Tags:      opts.Tags,
	}

	for _, ref := range opts.References {
		switch ref.Type {
		case ot.ChildOfRef, ot.FollowsFromRef:
			parentCtx := ref.ReferencedContext.(SpanContext)
			span.context = NewSpanContext(parentCtx)

			return span
		}
	}

	span.context = NewRootSpanContext()
	span.context.Sampled = r.options.ShouldSample(span.context.TraceID)

	return span
}

func shouldSample(traceID int64) bool {
	return false
}

// NewTracer Get a new Tracer with the default options applied.
func NewTracer() ot.Tracer {
	return NewTracerWithOptions(&Options{})
}

// NewTracerWithOptions Get a new Tracer with the specified options.
func NewTracerWithOptions(options *Options) ot.Tracer {
	return NewTracerWithEverything(options, NewRecorder())
}

// NewTracerWithEverything Get a new Tracer with the works.
func NewTracerWithEverything(options *Options, recorder SpanRecorder) ot.Tracer {
	InitSensor(options)
	ret := &tracerS{
		options: TracerOptions{
			Recorder:       recorder,
			ShouldSample:   shouldSample,
			MaxLogsPerSpan: MaxLogsPerSpan,
		},
	}

	return ret
}
