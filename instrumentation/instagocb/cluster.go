// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"time"

	"github.com/couchbase/gocb/v2"
)

type Cluster interface {
	Bucket(bucketName string) Bucket
	Buckets() *gocb.BucketManager
	WaitUntilReady(timeout time.Duration, opts *gocb.WaitUntilReadyOptions) error
	Close(opts *gocb.ClusterCloseOptions) error
	Users() *gocb.UserManager
	AnalyticsIndexes() *gocb.AnalyticsIndexManager
	QueryIndexes() *gocb.QueryIndexManager
	SearchIndexes() *gocb.SearchIndexManager
	EventingFunctions() *gocb.EventingFunctionManager
	Transactions() *gocb.Transactions
}

type InstanaCluster struct {
	iTracer requestTracer
	*gocb.Cluster
}

func (ic *InstanaCluster) Bucket(bn string) Bucket {
	b := &InstanaBucket{
		iTracer: ic.iTracer,
		Bucket:  ic.Cluster.Bucket(bn),
	}
	return b
}
