// (c) Copyright IBM Corp. 2023

package instagraphql

import (
	"context"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

func removeHTTPTags(sp ot.Span) {
	sp.SetTag("http.route_id", nil)
	sp.SetTag("http.method", nil)
	sp.SetTag("http.protocol", nil)
	sp.SetTag("http.host", nil)
	sp.SetTag("http.path", nil)
	sp.SetTag("http.header", nil)
}

func instrument(ctx context.Context, sensor *instana.Sensor, p *graphql.Params, res *graphql.Result, isSubscribe bool) *graphql.Result {
	var sp ot.Span
	var ok bool

	if sp, ok = instana.SpanFromContext(ctx); ok {
		// We repurpose the http span to become a GraphQL span. This way we trace only one entry span instead of two
		sp.SetOperationName("graphql.server")

		// Remove http tags from the span to guarantee that the repurposed span will behave accordingly
		removeHTTPTags(sp)
	} else {
		t := sensor.Tracer()

		if isSubscribe {
			opts := []ot.StartSpanOption{
				ext.SpanKindRPCClient,
			}

			sp = t.StartSpan("graphql.client", opts...)
		} else {
			sp = t.StartSpan("graphql.server")
		}

		// The GraphQL span is supposed to always be related to an HTTP parent span.
		// If for whatever reason there was not a parent HTTP span, we finish the GraphQL span.
		// Otherwise, we leave it to the HTTP span to be finished when done.
		defer sp.Finish()
	}

	dt, err := parseQuery(p.RequestString)

	if res == nil {
		res = graphql.Do(*p)
	}

	if err != nil {
		sp.SetTag("graphql.error", err.Error())
		sp.LogFields(otlog.Object("error", err))

		return res
	}

	sp.SetTag("graphql.operationType", dt.opType)
	sp.SetTag("graphql.operationName", dt.opName)
	sp.SetTag("graphql.fields", dt.fieldMap)
	sp.SetTag("graphql.args", dt.argMap)

	if len(res.Errors) > 0 {
		var err []string

		for _, e := range res.Errors {
			err = append(err, e.Error())
		}

		sp.SetTag("graphql.error", strings.Join(err, ", "))
		sp.LogFields(otlog.Object("error", err))
	}

	return res
}

// Do wraps the original graphql.Do, traces the GraphQL query and returns the result of the original graphql.Do
func Do(ctx context.Context, sensor *instana.Sensor, p graphql.Params) *graphql.Result {
	return instrument(ctx, sensor, &p, nil, false)
}

// Subscribe wraps the original graphql.Subscribe, traces the GraphQL query and returns the result of the original graphql.Subscribe
func Subscribe(ctx context.Context, sensor *instana.Sensor, p graphql.Params) chan *graphql.Result {
	originalCh := graphql.Subscribe(p)
	ch := make(chan *graphql.Result, len(originalCh))

	go func() {
	loop:
		for {
			select {
			case res, isOpen := <-originalCh:
				if !isOpen {
					close(ch)
					break loop
				}

				_ = instrument(ctx, sensor, &p, res, true)

				ch <- res

			case <-ctx.Done():
				break loop
			}
		}
	}()

	return ch
}

// ResultCallbackFn traces the GraphQL query and executes the original handler.ResultCallbackFn if fn is provided.
func ResultCallbackFn(sensor *instana.Sensor, fn handler.ResultCallbackFn) handler.ResultCallbackFn {
	return func(ctx context.Context, p *graphql.Params, res *graphql.Result, responseBody []byte) {
		_ = instrument(ctx, sensor, p, res, false)

		if fn != nil {
			fn(ctx, p, res, responseBody)
		}
	}
}
