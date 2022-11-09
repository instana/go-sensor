// (c) Copyright IBM Corp. 2022

package instana

import (
	"net/url"
	"strings"
)

const maxDelayedSpans = 500

var delayed = &delayedSpans{
	spans: make(chan *spanS, maxDelayedSpans),
}

type delayedSpans struct {
	spans chan *spanS
}

func (ds *delayedSpans) append(span *spanS) bool {
	select {
	case ds.spans <- span:
		return true
	default:
		return false
	}
}

func (ds *delayedSpans) process(stop chan struct{}) <-chan *spanS {
	c := make(chan *spanS)

	go func() {
		defer close(c)
		var s *spanS

		for {
			select {
			case <-stop:
				return
			default:
				select {
				case s = <-ds.spans:
					if t, ok := s.Tracer().(Tracer); ok {
						opts := t.Options()

						newParams := url.Values{}
						if httpParamsI, ok := s.Tags["http.params"]; ok {
							if httpParams, ok := httpParamsI.(string); ok {
								p := strings.Split(httpParams, "&")
								for _, pair := range p {
									kv := strings.Split(pair, "=")

									if len(kv) == 2 {
										key, value := kv[0], kv[1]
										if opts.Secrets.Match(key) {
											newParams.Set(key, "<redacted>")
										} else {
											newParams.Set(key, value)
										}
									}
								}
							}
						}

						if len(newParams) > 0 {
							s.SetTag("http.params", newParams.Encode())
						}

						c <- s
					}
				default:
					return
				}
			}

		}

	}()

	return c
}

func (ds *delayedSpans) flush() {
	stop := make(chan struct{}, 1)
	c := ds.process(stop)
	for {
		s, ok := <-c
		agentReady := sensor.Agent().Ready()
		if ok && agentReady {
			s.tracer.recorder.RecordSpan(s)

			continue
		}

		if !agentReady || !ok {
			stop <- struct{}{}

			if s != nil {
				ds.append(s)
			}

			for spanToDelayAgain := range c {
				ds.append(spanToDelayAgain)
			}

			break
		}
	}
}
