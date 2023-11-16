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
	Transactions() *gocb.Transactions

	// search query
	SearchQuery(indexName string, query cbsearch.Query, opts *gocb.SearchOptions) (*gocb.SearchResult, error)

	// query
	Query(statement string, opts *gocb.QueryOptions) (*gocb.QueryResult, error)

	// ping
	Ping(opts *gocb.PingOptions) (*gocb.PingResult, error)

	// internal
	Internal() *gocb.InternalCluster

	// diag
	Diagnostics(opts *gocb.DiagnosticsOptions) (*gocb.DiagnosticsResult, error)

	//analytics query
	AnalyticsQuery(statement string, opts *gocb.AnalyticsOptions) (*gocb.AnalyticsResult, error)
}

type InstanaCluster struct {
	iTracer requestTracer
	*gocb.Cluster
}

// Bucket connects the cluster to server(s) and returns a new Bucket instance.
func (ic *InstanaCluster) Bucket(bucketName string) Bucket {
	bucket := ic.Cluster.Bucket(bucketName)
	return createBucket(ic.iTracer, bucket)
}

// Buckets returns a BucketManager for managing buckets.
func (ic *InstanaCluster) Buckets() BucketManager {
	bm := ic.Cluster.Buckets()
	return createBucketManager(ic.iTracer, bm)

}
