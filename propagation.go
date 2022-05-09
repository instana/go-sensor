// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"errors"
	"net/http"
	"strings"

	"github.com/instana/go-sensor/w3ctrace"
	ot "github.com/opentracing/opentracing-go"
)

// Instana header constants
const (
	// FieldT Trace ID header
	FieldT = "x-instana-t"
	// FieldS Span ID header
	FieldS = "x-instana-s"
	// FieldL Level header
	FieldL = "x-instana-l"
	// FieldB OT Baggage header
	FieldB = "x-instana-b-"
	// FieldSynthetic if set to 1, marks the call as synthetic, e.g.
	// a healthcheck request
	FieldSynthetic = "x-instana-synthetic"
)

func injectTraceContext(sc SpanContext, opaqueCarrier interface{}) error {
	roCarrier, ok := opaqueCarrier.(ot.TextMapReader)
	if !ok {
		return ot.ErrInvalidCarrier
	}

	// Handle pre-existing case-sensitive keys
	exstfieldT := FieldT
	exstfieldS := FieldS
	exstfieldL := FieldL
	exstfieldB := FieldB

	roCarrier.ForeachKey(func(k, v string) error {
		switch strings.ToLower(k) {
		case FieldT:
			exstfieldT = k
		case FieldS:
			exstfieldS = k
		case FieldL:
			exstfieldL = k
		default:
			if strings.HasPrefix(strings.ToLower(k), FieldB) {
				exstfieldB = string([]rune(k)[:len(FieldB)])
			}
		}
		return nil
	})

	carrier, ok := opaqueCarrier.(ot.TextMapWriter)
	if !ok {
		return ot.ErrInvalidCarrier
	}

	if c, ok := opaqueCarrier.(ot.HTTPHeadersCarrier); ok {
		// Even though the godoc claims that the key passed to (*http.Header).Set()
		// is case-insensitive, it actually normalizes it using textproto.CanonicalMIMEHeaderKey()
		// before populating the value. As a result headers with non-canonical will not be
		// overwritten with a new value. This is only the case if header names were set while
		// initializing the http.Header instance, i.e.
		//     h := http.Headers{"X-InStAnA-T": {"abc123"}}
		// and does not apply to a common case when requests are being created using http.NewRequest()
		// or http.ReadRequest() that call (*http.Header).Set() to set header values.
		h := http.Header(c)
		delete(h, exstfieldT)
		delete(h, exstfieldS)
		delete(h, exstfieldL)

		for key := range h {
			if strings.HasPrefix(strings.ToLower(key), FieldB) {
				delete(h, key)
			}
		}

		addW3CTraceContext(h, sc)
		addEUMHeaders(h, sc)
	}

	if !sc.Suppressed {
		carrier.Set(exstfieldT, FormatID(sc.TraceID))
		carrier.Set(exstfieldS, FormatID(sc.SpanID))
	} else {
		// remove trace context keys from the carrier
		switch c := opaqueCarrier.(type) {
		case ot.HTTPHeadersCarrier:
			h := http.Header(c)
			h.Del(exstfieldT)
			h.Del(exstfieldS)
		case ot.TextMapCarrier:
			delete(c, exstfieldT)
			delete(c, exstfieldS)
		case interface{ RemoveAll() }:
			// in case carrier has the RemoveAll() method that wipes all trace
			// headers, for example the instasarama.ProducerMessagCarrier, we
			// use it to remove the context of a suppressed trace
			c.RemoveAll()
		}
	}

	carrier.Set(exstfieldL, formatLevel(sc))

	for k, v := range sc.Baggage {
		carrier.Set(exstfieldB+k, v)
	}

	return nil
}

// This method searches for Instana headers (FieldT, FieldS, FieldL and header with name prefixed with FieldB)
// and try to parse their values. It also tries to extract W3C context and assign it inside returned object. W3C context
// will be propagated further and can be used as a fallback.
func extractTraceContext(opaqueCarrier interface{}) (SpanContext, error) {
	spanContext := SpanContext{
		Baggage: make(map[string]string),
	}

	carrier, ok := opaqueCarrier.(ot.TextMapReader)
	if !ok {
		return spanContext, ot.ErrInvalidCarrier
	}

	if c, ok := opaqueCarrier.(ot.HTTPHeadersCarrier); ok {
		pickupW3CTraceContext(http.Header(c), &spanContext)
	}

	// Iterate over the headers, look for Instana headers and try to parse them.
	// In case of error interrupt, iteration and return it.
	err := carrier.ForeachKey(func(k, v string) error {
		var err error

		switch strings.ToLower(k) {
		case FieldT:
			spanContext.TraceIDHi, spanContext.TraceID, err = ParseLongID(v)
			if err != nil {
				return ot.ErrSpanContextCorrupted
			}
		case FieldS:
			spanContext.SpanID, err = ParseID(v)
			if err != nil {
				return ot.ErrSpanContextCorrupted
			}
		case FieldL:
			// When FieldL is present and equals to "0", then spanContext is suppressed.
			// In addition to that non-empty correlation data may be extracted.
			suppressed, corrData, err := parseLevel(v)
			if err != nil {
				sensor.logger.Info("failed to parse ", k, ": ", err, " (", v, ")")
				// use defaults
				suppressed, corrData = false, EUMCorrelationData{}
			}

			spanContext.Suppressed = suppressed
			if !spanContext.Suppressed {
				spanContext.Correlation = corrData
			}
		default:
			if strings.HasPrefix(strings.ToLower(k), FieldB) {
				// preserve original case of the baggage key
				spanContext.Baggage[k[len(FieldB):]] = v
			}
		}

		return nil
	})
	if err != nil {
		return spanContext, err
	}

	// reset the trace IDs if a correlation ID has been provided
	if spanContext.Correlation.ID != "" {
		spanContext.TraceIDHi, spanContext.TraceID, spanContext.SpanID = 0, 0, 0

		return spanContext, nil
	}

	if spanContext.IsZero() {
		return spanContext, ot.ErrSpanContextNotFound
	}

	// When the context is not suppressed and one of Instana ID headers set.
	if !spanContext.Suppressed &&
		(spanContext.SpanID == 0 != (spanContext.TraceIDHi == 0 && spanContext.TraceID == 0)) {
		sensor.logger.Debug("broken Instana trace context:",
			" SpanID=", FormatID(spanContext.SpanID),
			" TraceID=", FormatLongID(spanContext.TraceIDHi, spanContext.TraceID))

		// Check if w3 context was found
		if !spanContext.W3CContext.IsZero() {
			// Return SpanContext with w3 context, ignore other values
			return SpanContext{
				W3CContext: spanContext.W3CContext,
			}, nil
		}

		return spanContext, ot.ErrSpanContextCorrupted
	}

	return spanContext, nil
}

func addW3CTraceContext(h http.Header, sc SpanContext) {
	traceID, spanID := FormatLongID(sc.TraceIDHi, sc.TraceID), FormatID(sc.SpanID)
	trCtx := sc.W3CContext

	// check for an existing w3c trace
	if trCtx.IsZero() {
		// initiate trace if none
		trCtx = w3ctrace.New(w3ctrace.Parent{
			Version:  w3ctrace.Version_Max,
			TraceID:  traceID,
			ParentID: spanID,
			Flags: w3ctrace.Flags{
				Sampled: !sc.Suppressed,
			},
		})
	}

	// update the traceparent parent ID
	p := trCtx.Parent()
	p.ParentID = spanID
	// sync the traceparent `sampled` flags with the X-Instana-L value
	p.Flags.Sampled = !sc.Suppressed

	trCtx.RawParent = p.String()

	// participate in w3c trace context if tracing is enabled
	if !sc.Suppressed {
		// propagate truncated trace ID downstream
		trCtx.RawState = w3ctrace.FormStateWithInstanaTraceStateValue(trCtx.State(), FormatID(sc.TraceID)+";"+spanID).String()
	}

	w3ctrace.Inject(trCtx, h)
}

func pickupW3CTraceContext(h http.Header, sc *SpanContext) {
	trCtx, err := w3ctrace.Extract(h)
	if err != nil {
		return
	}
	sc.W3CContext = trCtx
}

func addEUMHeaders(h http.Header, sc SpanContext) {
	// Preserve original Server-Timing header values by combining them into a comma-separated list
	st := append(h["Server-Timing"], "intid;desc="+FormatID(sc.TraceID))
	h.Set("Server-Timing", strings.Join(st, ", "))
}

var errMalformedHeader = errors.New("malformed header value")

func parseLevel(s string) (bool, EUMCorrelationData, error) {
	const (
		levelState uint8 = iota
		partSeparatorState
		correlationPartState
		correlationTypeState
		correlationIDState
		finalState
	)

	if s == "" {
		return false, EUMCorrelationData{}, nil
	}

	var (
		typeInd                 int
		state                   uint8
		level, corrType, corrID string
	)
PARSE:
	for ptr := 0; state != finalState && ptr < len(s); ptr++ {
		switch state {
		case levelState: // looking for 0 or 1
			level = s[ptr : ptr+1]

			if level != "0" && level != "1" {
				break PARSE
			}

			if ptr == len(s)-1 { // no correlation ID provided
				state = finalState
			} else {
				state = partSeparatorState
			}
		case partSeparatorState: // skip OWS while looking for ','
			switch s[ptr] {
			case ' ', '\t': // advance
			case ',':
				state = correlationPartState
			default:
				break PARSE
			}
		case correlationPartState: // skip OWS while searching for 'correlationType=' prefix
			switch {
			case s[ptr] == ' ' || s[ptr] == '\t': // advance
			case strings.HasPrefix(s[ptr:], "correlationType="):
				ptr += 15 // advance to the end of prefix
				typeInd = ptr + 1
				state = correlationTypeState
			default:
				break PARSE
			}
		case correlationTypeState: // skip OWS while looking for ';'
			switch s[ptr] {
			case ' ', '\t': // possibly trailing OWS, advance
			case ';':
				state = correlationIDState
			default:
				corrType = s[typeInd : ptr+1]
			}
		case correlationIDState: //  skip OWS while searching for 'correlationId=' prefix
			switch {
			case s[ptr] == ' ' || s[ptr] == '\t': // leading OWS, advance
			case strings.HasPrefix(s[ptr:], "correlationId="):
				ptr += 14
				corrID = s[ptr:]
				state = finalState
			default:
				break PARSE
			}
		default:
			break PARSE
		}
	}

	if state != finalState {
		return false, EUMCorrelationData{}, errMalformedHeader
	}

	return level == "0", EUMCorrelationData{Type: corrType, ID: corrID}, nil
}

func formatLevel(sc SpanContext) string {
	if sc.Suppressed {
		return "0"
	}

	return "1"
}
