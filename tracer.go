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

	sc := NewRootSpanContext()
	for _, ref := range opts.References {
		if ref.Type == ot.ChildOfRef || ref.Type == ot.FollowsFromRef {
			sc = NewSpanContext(ref.ReferencedContext.(SpanContext))
			break
		}
	}

	if tag, ok := opts.Tags[suppressTracingTag]; ok {
		sc.Suppressed = tag.(bool)
		delete(opts.Tags, suppressTracingTag)
	}

	sc.Sampled = r.options.ShouldSample(sc.TraceID)

	return &spanS{
		context:   sc,
		tracer:    r,
		Service:   sensor.serviceName,
		Operation: operationName,
		Start:     startTime,
		Duration:  -1,
		Tags:      opts.Tags,
	}
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
