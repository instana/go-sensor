package instana

import (
	"net/http"
	"strconv"
	"strings"

	ot "github.com/opentracing/opentracing-go"
)

type textMapPropagator struct {
	tracer *tracerS
}

// Instana header constants
const (
	// FieldT Trace ID header
	FieldT = "x-instana-t"
	// FieldS Span ID header
	FieldS = "x-instana-s"
	// FieldParentSpanID Parent Span ID header
	FieldParentSpanID = "x-instana-parentspanid"
	// FieldL Level header
	FieldL = "x-instana-l"
	// FieldB OT Baggage header
	FieldB     = "x-instana-b-"
	fieldCount = 2
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
		exstfieldT            = FieldT
		exstfieldS            = FieldS
		exstfieldParentSpanID = FieldParentSpanID
		exstfieldL            = FieldL
		exstfieldB            = FieldB
	)

	roCarrier.ForeachKey(func(k, v string) error {
		switch strings.ToLower(k) {
		case FieldT:
			exstfieldT = k
		case FieldS:
			exstfieldS = k
		case exstfieldParentSpanID:
			exstfieldParentSpanID = k
		case FieldL:
			exstfieldL = k
		default:
			if strings.HasPrefix(strings.ToLower(k), FieldB) {
				exstfieldB = string([]rune(k)[0:len(FieldB)])
			}
		}
		return nil
	})

	carrier, ok := opaqueCarrier.(ot.TextMapWriter)
	if !ok {
		return ot.ErrInvalidCarrier
	}

	hhcarrier, ok := opaqueCarrier.(ot.HTTPHeadersCarrier)
	if ok {
		// If http.Headers has pre-existing keys, calling Set() like we do
		// below will just append to those existing values and break context
		// propagation.  So defend against that case, we delete any pre-existing
		// keys entirely first.
		y := http.Header(hhcarrier)
		y.Del(exstfieldT)
		y.Del(exstfieldS)
		y.Del(exstfieldParentSpanID)
		y.Del(exstfieldL)

		for key := range y {
			if strings.HasPrefix(strings.ToLower(key), FieldB) {
				y.Del(key)
			}
		}
	}

	if instanaTID, err := ID2Header(sc.TraceID); err == nil {
		carrier.Set(exstfieldT, instanaTID)
	} else {
		log.debug(err)
	}
	if instanaSID, err := ID2Header(sc.SpanID); err == nil {
		carrier.Set(exstfieldS, instanaSID)
	} else {
		log.debug(err)
	}

	if sc.ParentSpanID != 0 {
		if instanaParentSpanID, err := ID2Header(sc.ParentSpanID); err == nil {
			carrier.Set(exstfieldParentSpanID, instanaParentSpanID)
		} else {
			log.debug(err)
		}
	}

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
	var traceID, spanID, parentSpanID int64
	var err error
	baggage := make(map[string]string)
	err = carrier.ForeachKey(func(k, v string) error {
		switch strings.ToLower(k) {
		case FieldT:
			fieldCount++
			traceID, err = Header2ID(v)
			if err != nil {
				return ot.ErrSpanContextCorrupted
			}
		case FieldS:
			fieldCount++
			spanID, err = Header2ID(v)
			if err != nil {
				return ot.ErrSpanContextCorrupted
			}
		case FieldParentSpanID:
			fieldCount++
			parentSpanID, err = Header2ID(v)
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

	if fieldCount == 0 {
		return nil, ot.ErrSpanContextNotFound
	} else if fieldCount < 2 {
		return nil, ot.ErrSpanContextCorrupted
	} else if err != nil {
		return nil, err
	}

	return SpanContext{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		Sampled:      false,
		Baggage:      baggage,
	}, nil
}
