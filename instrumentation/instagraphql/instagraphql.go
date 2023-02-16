// (c) Copyright IBM Corp. 2023

package instagraphql

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var mu sync.RWMutex

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
	mt := p.Schema.MutationType().Fields()

	for k := range dt.fieldMap {
		if mutType, ok := mt[k]; ok {
			mu.Lock()
			mutationSpans[mutType.Type.Name()] = sp
			mu.Unlock()

			// cleanup the map after one second
			go func(n string) {
				time.Sleep(time.Second)

				mu.Lock()
				delete(mutationSpans, n)
				mu.Unlock()
			}(mutType.Type.Name())

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
		// defer sp.Finish()
	}

	return // span, repurposed
}

// mutationSpans is a map of spans originated from mutations where the key is the object type name set in the schema.
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
var mutationSpans = make(map[string]ot.Span)

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

	res := graphql.Do(*p)

	dt, err := parseQuery(p.RequestString)

	if err != nil {
		sp.SetTag("graphql.error", err.Error())
		sp.LogFields(otlog.Object("error", err))

		return res
	}

	if dt.opType == "mutation" {
		cacheMutationSpan(sp, dt, p)
	}

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
		// We lookup for a matching key in mutationSpan for a few millisencods.
		// This is needed because the subscription runs in a go routine and may be triggered before the mutationSpans
		// map received a reference to the mutation span
		for i := 0; i < 3; i++ {
			time.Sleep(time.Millisecond * 1)
			mu.RLock()
			ps = mutationSpans[mutKey]
			mu.RUnlock()
			if ps != nil {
				break
			}
		}
	}

	if ps != nil {
		opts := []ot.StartSpanOption{
			ext.SpanKindRPCClient,
		}

		opts = append(opts, ot.ChildOf(ps.Context()))
		sp = ps.Tracer().StartSpan("graphql.client", opts...)
	}

	feedTags(sp, dt, res)

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
