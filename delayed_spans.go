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
						if paramsTag, ok := s.Tags["http.params"]; ok {
							if httpParams, ok := paramsTag.(string); ok {
								p, err := url.ParseQuery(httpParams)
								if err != nil {
									continue
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

		if !agentReady {
			stop <- struct{}{}
		}

		if !ok {
			break
		}

		// put back in the queue
		ds.append(s)

		for spanToDelayAgain := range c {
			ds.append(spanToDelayAgain)
		}

		break

	}
}
