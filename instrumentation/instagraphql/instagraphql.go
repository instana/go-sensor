// (c) Copyright IBM Corp. 2023

package instagraphql

import (
	"context"
	"strings"

	"github.com/graphql-go/graphql"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
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

func Do(ctx context.Context, sensor *instana.Sensor, p graphql.Params) *graphql.Result {
	var sp ot.Span
	var ok bool

	if sp, ok = instana.SpanFromContext(ctx); ok {
		// We repurpose the http span to become a GraphQL span. This way we trace only one entry span instead of two
		sp.SetOperationName("graphql.server")

		// Remove http tags from the span to guarantee that the repurposed span will behave accordingly
		removeHTTPTags(sp)
	} else {
		t := sensor.Tracer()
		sp = t.StartSpan("graphql.server")

		// The GraphQL span is supposed to always be related to an HTTP parent span.
		// If for whatever reason there was not a parent HTTP span, we finish the GraphQL span.
		// Otherwise, we leave it to the HTTP span to be finished when done.
		defer sp.Finish()
	}

	dt, err := parseQuery(p.RequestString)

	if err != nil {
		sp.SetTag("graphql.error", err.Error())
		sp.LogFields(otlog.Object("error", err))

		return graphql.Do(p)
	}

	sp.SetTag("graphql.operationType", dt.opType)
	sp.SetTag("graphql.operationName", dt.opName)
	sp.SetTag("graphql.fields", dt.fieldMap)
	sp.SetTag("graphql.args", dt.argMap)

	res := graphql.Do(p)

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
