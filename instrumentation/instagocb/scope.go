// (c) Copyright IBM Corp. 2023

package instagocb

import "github.com/couchbase/gocb/v2"

type Scope interface {
	Name() string
	BucketName() string
	Collection(collectionName string) *gocb.Collection
	Query(statement string, opts *gocb.QueryOptions) (*gocb.QueryResult, error)
}

type InstanaScope struct {
	iTracer gocb.RequestTracer
	*gocb.Scope
}

func (is *InstanaScope) Query(statement string, opts *gocb.QueryOptions) (*gocb.QueryResult, error) {
	span := is.iTracer.RequestSpan(opts.ParentSpan.Context(), "instana-query")
	span.SetAttribute("db.statement", statement)
	span.SetAttribute("db.name", is.BucketName())
	span.SetAttribute("db.couchbase.scope", is.Name())

	res, err := is.Scope.Query(statement, opts)

	span.(*Span).err = err

	defer span.End()

	return res, err
}
