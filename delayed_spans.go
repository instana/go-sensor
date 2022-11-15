// (c) Copyright IBM Corp. 2022

package instana

import (
	"net/url"
)

const maxDelayedSpans = 500

var delayed = &delayedSpans{
	spans: make(chan *spanS, maxDelayedSpans),
}

type delayedSpans struct {
	spans chan *spanS
}

// append add a span to the buffer if buffer is not full yet
func (ds *delayedSpans) append(span *spanS) bool {
	select {
	case ds.spans <- span:
		return true
	default:
		return false
	}
}

// flush processes buffered spans and move them from the delayed buffer to the recorder if agent is ready
func (ds *delayedSpans) flush() {
	for {
		select {
		case s := <-ds.spans:
			t, ok := s.Tracer().(Tracer)
			if !ok {
				sensor.logger.Debug("span tracer has unexpected type")
				continue
			}

			if err := ds.processSpan(s, t.Options()); err != nil {
				sensor.logger.Debug("error while processing spans:", err.Error())
				continue
			}

			if sensor.Agent().Ready() {
				s.tracer.recorder.RecordSpan(s)
			} else {
				ds.append(s)
				return
			}
		default:
			return
		}
	}
}

// processSpan applies secret filtering to the buffered http span http.params tag
func (ds *delayedSpans) processSpan(s *spanS, opts TracerOptions) error {
	newParams := url.Values{}
	if paramsTag, ok := s.Tags["http.params"]; ok {
		if httpParams, ok := paramsTag.(string); ok {
			p, err := url.ParseQuery(httpParams)
			if err != nil {
				return err
			}

			for key, value := range p {
				if opts.Secrets.Match(key) {
					newParams[key] = []string{"<redacted>"}
				} else {
					newParams[key] = value
				}
			}
		}
	}

	if len(newParams) > 0 {
		s.SetTag("http.params", newParams.Encode())
	}

	return nil
}
