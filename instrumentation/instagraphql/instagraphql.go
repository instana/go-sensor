// (c) Copyright IBM Corp. 2023

package instagraphql

import (
	"context"

	"github.com/graphql-go/graphql"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	// ot "github.com/opentracing/opentracing-go"
	// otlog "github.com/opentracing/opentracing-go/log"
)

func Do(ctx context.Context, p graphql.Params) *graphql.Result {
	dt := parseQuery(p.RequestString)

	var sp ot.Span
	var ok bool

	if sp, ok = instana.SpanFromContext(ctx); ok {
		// TODO: check if parent span is actually an HTTP span
		// Remove http tags from the span to guarantee that the repurposed span will behave accordingly
		sp.SetTag("http.route_id", nil)
		sp.SetTag("http.method", nil)
		sp.SetTag("http.protocol", nil)
		sp.SetTag("http.host", nil)
		sp.SetTag("http.path", nil)
		sp.SetTag("http.header", nil)
	} else {
		sp = ot.StartSpan("graphql.server")
	}

	sp.SetOperationName("graphql.server")
	sp.SetTag("graphql.operationType", dt.opType)
	sp.SetTag("graphql.operationName", dt.opName)
	sp.SetTag("graphql.fields", dt.fieldMap)
	sp.SetTag("graphql.args", dt.argMap)

	// The GraphQL span is supposed to always be related to an HTTP parent span.
	// If for whatever reason there was not a parent HTTP span, we finish the GraphQL span.
	// Otherwise, we leave it to the HTTP span to be finished when done.
	if !ok {
		sp.Finish()
	}

	// todo: check for errors and add them to the span

	return graphql.Do(p)
}
