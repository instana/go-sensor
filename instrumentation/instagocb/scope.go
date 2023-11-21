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

type instaScope struct {
	iTracer gocb.RequestTracer
	*gocb.Scope
}

// Query executes the query statement on the server, constraining the query to the bucket and scope.
func (is *instaScope) Query(statement string, opts *gocb.QueryOptions) (*gocb.QueryResult, error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := is.iTracer.RequestSpan(tracectx, "QUERY")
	span.SetAttribute(operationSpanTag, statement)
	span.SetAttribute(bucketNameSpanTag, is.BucketName())

	res, err := is.Scope.Query(statement, opts)

	span.(*Span).err = err

	defer span.End()

	return res, err
}

// Collection returns an instance of a collection.
func (is *instaScope) Collection(collectionName string) Collection {
	collection := is.Scope.Collection(collectionName)
	return createCollection(is.iTracer, collection)

}

// Unwrap returns the original *gocb.Scope instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (is *instaScope) Unwrap() *gocb.Scope {
	return is.Scope
}

// helper functions

// createScope will wrap *gocb.Scope in to instanaScope and will return it as Scope interface
func createScope(tracer gocb.RequestTracer, scope *gocb.Scope) Scope {
	return &instaScope{
		iTracer: tracer,
		Scope:   scope,
	}
}
