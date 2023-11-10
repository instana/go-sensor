// (c) Copyright IBM Corp. 2023

package instagocb

import "github.com/couchbase/gocb/v2"

type Scope interface {
	Name() string
	BucketName() string
	Collection(collectionName string) Collection

	// query
	Query(statement string, opts *gocb.QueryOptions) (*gocb.QueryResult, error)

	// analytic query
	AnalyticsQuery(statement string, opts *gocb.AnalyticsOptions) (*gocb.AnalyticsResult, error)
}

type InstanaScope struct {
	iTracer gocb.RequestTracer
	*gocb.Scope
}

// Query executes the query statement on the server, constraining the query to the bucket and scope.
func (is *InstanaScope) Query(statement string, opts *gocb.QueryOptions) (*gocb.QueryResult, error) {
	span := is.iTracer.RequestSpan(opts.ParentSpan.Context(), "query")
	span.SetAttribute("db.statement", statement)
	span.SetAttribute("db.name", is.BucketName())
	span.SetAttribute("db.couchbase.scope", is.Name())

	res, err := is.Scope.Query(statement, opts)

	span.(*Span).err = err

	defer span.End()

	return res, err
}

// Collection returns an instance of a collection.
func (is *InstanaScope) Collection(collectionName string) Collection {
	collection := is.Scope.Collection(collectionName)
	return createCollection(is.iTracer, collection)

}

// helper functions

// createScope will wrap *gocb.Scope in to instanaScope and will return it as Scope interface
func createScope(tracer gocb.RequestTracer, scope *gocb.Scope) Scope {
	return &InstanaScope{
		iTracer: tracer,
		Scope:   scope,
	}
}
