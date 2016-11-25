package instana

import (
	"strconv"
	"strings"

	"github.com/opentracing/basictracer-go"
	ot "github.com/opentracing/opentracing-go"
)

type textMapPropagator struct {
	tracer *tracerS
}

const (
	FIELD_COUNT = 2
	FIELD_T     = "x-instana-t"
	FIELD_S     = "x-instana-s"
	FIELD_L     = "x-instana-l"
	FIELD_B     = "x-instana-b-"
)

func (p *textMapPropagator) inject(spanContext ot.SpanContext, opaqueCarrier interface{}) error {
	sc, ok := spanContext.(basictracer.SpanContext)
	if !ok {
		return ot.ErrInvalidSpanContext
	}

	carrier, ok := opaqueCarrier.(ot.TextMapWriter)
	if !ok {
		return ot.ErrInvalidCarrier
	}

	carrier.Set(FIELD_T, strconv.FormatUint(sc.TraceID, 16))
	carrier.Set(FIELD_S, strconv.FormatUint(sc.SpanID, 16))
	carrier.Set(FIELD_L, strconv.Itoa(1))

	for k, v := range sc.Baggage {
		carrier.Set(FIELD_B+k, v)
	}

	return nil
}

func (p *textMapPropagator) extract(opaqueCarrier interface{}) (ot.SpanContext, error) {
	carrier, ok := opaqueCarrier.(ot.TextMapReader)
	if !ok {
		return nil, ot.ErrInvalidCarrier
	}

	fieldCount := 0
	var traceID, spanID uint64
	var err error
	baggage := make(map[string]string)
	err = carrier.ForeachKey(func(k, v string) error {
		switch strings.ToLower(k) {
		case FIELD_T:
			fieldCount++
			traceID, err = strconv.ParseUint(v, 16, 64)
			if err != nil {
				return ot.ErrSpanContextCorrupted
			}
		case FIELD_S:
			fieldCount++
			spanID, err = strconv.ParseUint(v, 16, 64)
			if err != nil {
				return ot.ErrSpanContextCorrupted
			}
		default:
			lk := strings.ToLower(k)

			if strings.HasPrefix(lk, FIELD_B) {
				baggage[strings.TrimPrefix(lk, FIELD_B)] = v
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if fieldCount < FIELD_COUNT {
		if fieldCount == 0 {
			return nil, ot.ErrSpanContextNotFound
		}

		return nil, ot.ErrSpanContextCorrupted
	}

	return basictracer.SpanContext{
		TraceID: traceID,
		SpanID:  spanID,
		Sampled: false, //TODO: add configurable sampling strategy
		Baggage: baggage,
	}, nil
}
