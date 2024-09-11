// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"time"

	"github.com/couchbase/gocb/v2"
	cbsearch "github.com/couchbase/gocb/v2/search"
)

type Cluster interface {
	Bucket(bucketName string) Bucket
	WaitUntilReady(timeout time.Duration, opts *gocb.WaitUntilReadyOptions) error
	Close(opts *gocb.ClusterCloseOptions) error
	Users() *gocb.UserManager
	Buckets() BucketManager
	AnalyticsIndexes() *gocb.AnalyticsIndexManager
	QueryIndexes() *gocb.QueryIndexManager
	SearchIndexes() *gocb.SearchIndexManager
	EventingFunctions() *gocb.EventingFunctionManager
	Transactions() Transactions

	SearchQuery(indexName string, query cbsearch.Query, opts *gocb.SearchOptions) (*gocb.SearchResult, error)

	Query(statement string, opts *gocb.QueryOptions) (*gocb.QueryResult, error)

	Ping(opts *gocb.PingOptions) (*gocb.PingResult, error)

	Internal() *gocb.InternalCluster

	Diagnostics(opts *gocb.DiagnosticsOptions) (*gocb.DiagnosticsResult, error)

	AnalyticsQuery(statement string, opts *gocb.AnalyticsOptions) (*gocb.AnalyticsResult, error)

	// These methods are only available in Cluster interface, not available in gocb.Cluster instance
	Unwrap() *gocb.Cluster
	WrapTransactionAttemptContext(tac *gocb.TransactionAttemptContext, parentSpan gocb.RequestSpan) TransactionAttemptContext
}

type instaCluster struct {
	iTracer requestTracer
	*gocb.Cluster
}

// Bucket connects the cluster to server(s) and returns a new Bucket instance.
func (ic *instaCluster) Bucket(bucketName string) Bucket {
	bucket := ic.Cluster.Bucket(bucketName)
	return createBucket(ic.iTracer, bucket)
}

// Buckets returns a BucketManager for managing buckets.
func (ic *instaCluster) Buckets() BucketManager {
	bm := ic.Cluster.Buckets()
	return createBucketManager(ic.iTracer, bm)

}

// Query executes the query statement on the server.
func (ic *instaCluster) Query(statement string, opts *gocb.QueryOptions) (*gocb.QueryResult, error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "QUERY")
	span.SetAttribute(operationSpanTag, statement)

	res, err := ic.Cluster.Query(statement, opts)

	span.(*Span).err = err

	defer span.End()

	return res, err
}

// SearchQuery executes the analytics query statement on the server.
func (ic *instaCluster) SearchQuery(indexName string, query cbsearch.Query, opts *gocb.SearchOptions) (*gocb.SearchResult, error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "SEARCH_QUERY")
	span.SetAttribute(operationSpanTag, "SEARCH "+indexName)

	res, err := ic.Cluster.SearchQuery(indexName, query, opts)

	span.(*Span).err = err

	defer span.End()

	return res, err
}

// AnalyticsQuery executes the analytics query statement on the server.
func (ic *instaCluster) AnalyticsQuery(statement string, opts *gocb.AnalyticsOptions) (*gocb.AnalyticsResult, error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "ANALYTICS_QUERY")
	span.SetAttribute(operationSpanTag, statement)

	res, err := ic.Cluster.AnalyticsQuery(statement, opts)

	span.(*Span).err = err

	defer span.End()

	return res, err
}

// Transactions returns a Transactions instance for performing transactions.
func (ic *instaCluster) Transactions() Transactions {
	return createTransactions(ic.iTracer, ic.Cluster.Transactions())
}

// Unwrap returns the original *gocb.Cluster instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (ic *instaCluster) Unwrap() *gocb.Cluster {
	return ic.Cluster
}

func (ic *instaCluster) WrapTransactionAttemptContext(tac *gocb.TransactionAttemptContext, parentSpan gocb.RequestSpan) TransactionAttemptContext {
	return createTransactionAttemptContext(ic.iTracer, tac, parentSpan)
}

func (ic *instaCluster) Close(opts *gocb.ClusterCloseOptions) error {
	return ic.Cluster.Close(opts)
}

// Helper functions

// createCluster will wrap *gocb.Cluster in to instaCluster and will return it as Cluster interface
func createCluster(tracer requestTracer, cluster *gocb.Cluster) Cluster {
	return &instaCluster{
		iTracer: tracer,
		Cluster: cluster,
	}
}
