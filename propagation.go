package instana

import (
	"strconv"
	"strings"

	ot "github.com/opentracing/opentracing-go"
)

type textMapPropagator struct {
	tracer *tracerS
}

const (
	fieldCount = 2
	fieldT     = "x-instana-t"
	fieldS     = "x-instana-s"
	fieldL     = "x-instana-l"
	fieldB     = "x-instana-b-"
)

func (r *textMapPropagator) inject(spanContext ot.SpanContext, opaqueCarrier interface{}) error {
	sc, ok := spanContext.(SpanContext)
	if !ok {
		return ot.ErrInvalidSpanContext
	}

	roCarrier, ok := opaqueCarrier.(ot.TextMapReader)
	if !ok {
		return ot.ErrInvalidCarrier
	}

	// Handle pre-existing case-sensitive keys
	var (
		exstfieldT = fieldT
		exstfieldS = fieldS
		exstfieldL = fieldL
		exstfieldB = fieldB
	)

	roCarrier.ForeachKey(func(k, v string) error {
		switch strings.ToLower(k) {
		case fieldT:
			exstfieldT = k
		case fieldS:
			exstfieldS = k
		case fieldL:
			exstfieldL = k
		default:
			if strings.HasPrefix(strings.ToLower(k), fieldB) {
				exstfieldB = string([]rune(k)[0:len(fieldB)])
			}
		}

		return nil
	})

	carrier, ok := opaqueCarrier.(ot.TextMapWriter)
	if !ok {
		return ot.ErrInvalidCarrier
	}

	carrier.Set(exstfieldT, strconv.FormatInt(sc.TraceID, 16))
	carrier.Set(exstfieldS, strconv.FormatInt(sc.SpanID, 16))
	carrier.Set(exstfieldL, strconv.Itoa(1))

	for k, v := range sc.Baggage {
		carrier.Set(exstfieldB+k, v)
	}

	return nil
}

func (r *textMapPropagator) extract(opaqueCarrier interface{}) (ot.SpanContext, error) {
	carrier, ok := opaqueCarrier.(ot.TextMapReader)
	if !ok {
		return nil, ot.ErrInvalidCarrier
	}

	fieldCount := 0
	var traceID, spanID int64
	var err error
	baggage := make(map[string]string)
	err = carrier.ForeachKey(func(k, v string) error {
		switch strings.ToLower(k) {
		case fieldT:
			fieldCount++
			traceID, err = strconv.ParseInt(v, 16, 64)
			if err != nil {
				return ot.ErrSpanContextCorrupted
			}
		case fieldS:
			fieldCount++
			spanID, err = strconv.ParseInt(v, 16, 64)
			if err != nil {
				return ot.ErrSpanContextCorrupted
			}
		default:
			lk := strings.ToLower(k)

			if strings.HasPrefix(lk, fieldB) {
				baggage[strings.TrimPrefix(lk, fieldB)] = v
			}
		}

		return nil
	})

	return r.finishExtract(err, fieldCount, traceID, spanID, baggage)
}

func (r *textMapPropagator) finishExtract(err error,
	fieldCount int,
	traceID int64,
	spanID int64,
	baggage map[string]string) (ot.SpanContext, error) {
	if err != nil {
		return nil, err
	}

	if fieldCount < fieldCount {
		if fieldCount == 0 {
			return nil, ot.ErrSpanContextNotFound
		}

		return nil, ot.ErrSpanContextCorrupted
	}

	return SpanContext{
		TraceID: traceID,
		SpanID:  spanID,
		Sampled: false, //TODO: add configurable sampling strategy
		Baggage: baggage,
	}, nil
}
