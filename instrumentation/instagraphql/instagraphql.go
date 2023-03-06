// (c) Copyright IBM Corp. 2023

package instagraphql

import (
	"context"
	"strings"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// mutationSpans is an expiring map of spans originated from mutations where the key is the object type name set in the schema.
// Example for the key "Character":
//
//	var characterType = graphql.NewObject(graphql.ObjectConfig{
//		Name: "Character",
//		Fields: ...
//	})
//
// This is our best guess at linking a mutation to a subscription, as they are by no means related via anything else.
// There will be also an obvious chance that a mutation is not the original parent of subscriptions, as we keep track
// of only one mutation (parent span) per type. But technically, this should not be an issue, as the mutation will still
// refer to the same type.
// var mutationSpans = make(map[string]ot.Span)
var mutationSpans = ExpiringMap{}

// retry returns a channel that is fulfilled after `wait` is reached or `fn` returns true.
// It retries after `every` time.
func retry(wait, every time.Duration, fn func() bool) {
	// attempts to resolve fn right away
	if fn() {
		return
	}

	done := make(chan struct{})
	timer := time.NewTimer(wait)

	go func() {
	loop:
		for {
			time.Sleep(every)
			select {
			case <-timer.C:
				done <- struct{}{}
				break loop
			default:
				if fn() {
					done <- struct{}{}
					break loop
				}
			}
		}
	}()

	<-done
}

func removeHTTPTags(sp ot.Span) {
	sp.SetTag("http.route_id", nil)
	sp.SetTag("http.method", nil)
	sp.SetTag("http.protocol", nil)
	sp.SetTag("http.host", nil)
	sp.SetTag("http.path", nil)
	sp.SetTag("http.header", nil)
}

func feedTags(sp ot.Span, dt *gqlData, res *graphql.Result) {
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
}

func cacheMutationSpan(sp ot.Span, dt *gqlData, p *graphql.Params) {
	fields := p.Schema.MutationType().Fields()

	for k := range dt.fieldMap {
		if mutType, ok := fields[k]; ok {
			mutationSpans.Set(mutType.Type.Name(), sp, time.Second)
			break
		}
	}
}

func extractSpan(ctx context.Context, sensor *instana.Sensor) (span ot.Span, repurposed bool) {
	if span, repurposed = instana.SpanFromContext(ctx); repurposed {
		// We repurpose the http span to become a GraphQL span. This way we trace only one entry span instead of two
		span.SetOperationName("graphql.server")

		// Remove http tags from the span to guarantee that the repurposed span will behave accordingly
		removeHTTPTags(span)
	} else {
		span = sensor.Tracer().StartSpan("graphql.server")
	}

	return // span, repurposed
}

func instrumentCallback(ctx context.Context, sensor *instana.Sensor, p *graphql.Params, res *graphql.Result) {
	var sp ot.Span
	var repurposed bool

	if sp, repurposed = extractSpan(ctx, sensor); !repurposed {
		defer sp.Finish()
	}

	dt, err := parseQuery(p.RequestString)

	if err != nil {
		sp.SetTag("graphql.error", err.Error())
		sp.LogFields(otlog.Object("error", err))

		return
	}

	if dt.opType == "mutation" {
		cacheMutationSpan(sp, dt, p)
	}

	feedTags(sp, dt, res)
}

func instrumentDo(ctx context.Context, sensor *instana.Sensor, p *graphql.Params) *graphql.Result {
	var sp ot.Span
	var repurposed bool

	if sp, repurposed = extractSpan(ctx, sensor); !repurposed {
		defer sp.Finish()
	}

	dt, err := parseQuery(p.RequestString)

	if err != nil {
		sp.SetTag("graphql.error", err.Error())
		sp.LogFields(otlog.Object("error", err))

		return graphql.Do(*p)
	}

	if dt.opType == "mutation" {
		cacheMutationSpan(sp, dt, p)
	}

	res := graphql.Do(*p)

	feedTags(sp, dt, res)

	return res
}

func instrumentSubscription(sensor *instana.Sensor, p *graphql.Params, res *graphql.Result) {
	var sp, ps ot.Span

	dt, err := parseQuery(p.RequestString)

	if err != nil {
		sp.SetTag("graphql.error", err.Error())
		sp.LogFields(otlog.Object("error", err))
		sp.Finish()
		return
	}

	t := sensor.Tracer()

	opts := []ot.StartSpanOption{
		ext.SpanKindRPCClient,
	}

	sp = t.StartSpan("graphql.client", opts...)

	subFields := p.Schema.SubscriptionType().Fields()

	// The key of dt.fieldMap should match a key in st, which should give us the name of the type being "mutated".
	// We will need this info to correlate the mutation to subscriptions.
	var mutKey string
	for k := range dt.fieldMap {
		if mutType, ok := subFields[k]; ok {
			mutKey = mutType.Type.Name()
			break
		}
	}

	if mutKey != "" {
		// Sometimes a mutation is triggered but a span was not yet created and the subscription span is
		// already "looking for" its parent. That's because these operations run concurrently.
		// We can retry to recover the belated mutation span for a while.
		// Eg: if the mutation is instrumented via ResultCallbackFn, the mutation has already happened when we start to
		// create the span, which could be caught concurrently by the subscription instrumentation.
		// Important: make sure to keep the wait short
		retry(time.Millisecond*100, time.Millisecond*1, func() bool {
			ps = mutationSpans.Get(mutKey)
			return ps != nil
		})
	}

	if ps != nil {
		opts := []ot.StartSpanOption{
			ext.SpanKindRPCClient,
		}

		opts = append(opts, ot.ChildOf(ps.Context()))
		sp = ps.Tracer().StartSpan("graphql.client", opts...)
	}

	feedTags(sp, dt, res)

	// "subscription-update" is used instead of "subscription" as the operation type.
	// It's a special case that renders all subscriptions grouped into one service called "GraphQL Subscribers" in the UI.
	sp.SetTag("graphql.operationType", "subscription-update")

	sp.Finish()
}

// Do wraps the original graphql.Do, traces the GraphQL query and returns the result of the original graphql.Do
func Do(ctx context.Context, sensor *instana.Sensor, p graphql.Params) *graphql.Result {
	return instrumentDo(ctx, sensor, &p)
}

// ResultCallbackFn traces the GraphQL query and executes the original handler.ResultCallbackFn if fn is provided.
func ResultCallbackFn(sensor *instana.Sensor, fn handler.ResultCallbackFn) handler.ResultCallbackFn {
	return func(ctx context.Context, p *graphql.Params, res *graphql.Result, responseBody []byte) {
		instrumentCallback(ctx, sensor, p, res)

		if fn != nil {
			fn(ctx, p, res, responseBody)
		}
	}
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
					break loop
				}

				instrumentSubscription(sensor, &p, res)

				ch <- res

			case <-ctx.Done():
				break loop
			}
		}
	}()

	return ch
}
