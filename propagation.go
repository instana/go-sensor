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
	FieldCount = 2
	FieldT     = "x-instana-t"
	FieldS     = "x-instana-s"
	FieldL     = "x-instana-l"
	FieldB     = "x-instana-b-"
)

func (r *textMapPropagator) inject(spanContext ot.SpanContext, opaqueCarrier interface{}) error {
	sc, ok := spanContext.(basictracer.SpanContext)
	if !ok {
		return ot.ErrInvalidSpanContext
	}

	carrier, ok := opaqueCarrier.(ot.TextMapWriter)
	if !ok {
		return ot.ErrInvalidCarrier
	}

	carrier.Set(FieldT, strconv.FormatUint(sc.TraceID, 16))
	carrier.Set(FieldS, strconv.FormatUint(sc.SpanID, 16))
	carrier.Set(FieldL, strconv.Itoa(1))

	for k, v := range sc.Baggage {
		carrier.Set(FieldB+k, v)
	}

	return nil
}

func (r *textMapPropagator) extract(opaqueCarrier interface{}) (ot.SpanContext, error) {
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
		case FieldT:
			fieldCount++
			traceID, err = strconv.ParseUint(v, 16, 64)
			if err != nil {
				return ot.ErrSpanContextCorrupted
			}
		case FieldS:
			fieldCount++
			spanID, err = strconv.ParseUint(v, 16, 64)
			if err != nil {
				return ot.ErrSpanContextCorrupted
			}
		default:
			lk := strings.ToLower(k)

			if strings.HasPrefix(lk, FieldB) {
				baggage[strings.TrimPrefix(lk, FieldB)] = v
			}
		}

		return nil
	})

	return r.finishExtract(err, fieldCount, traceID, spanID, baggage)
}

func (r *textMapPropagator) finishExtract(err error,
	fieldCount int,
	traceID uint64,
	spanID uint64,
	baggage map[string]string) (ot.SpanContext, error) {
	if err != nil {
		return nil, err
	}

	if fieldCount < FieldCount {
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
