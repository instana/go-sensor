// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"github.com/couchbase/gocb/v2"
)

type Transactions interface {
	Run(logicFn gocb.AttemptFunc, perConfig *gocb.TransactionOptions) (*gocb.TransactionResult, error)

	Unwrap() *gocb.Transactions
}

type instaTransactions struct {
	iTracer requestTracer
	*gocb.Transactions
}

// Run runs a lambda to perform a number of operations as part of a
// singular transaction.
func (it *instaTransactions) Run(logicFn gocb.AttemptFunc, perConfig *gocb.TransactionOptions) (*gocb.TransactionResult, error) {
	// bucket := ic.Cluster.Bucket(bucketName)
	return it.Transactions.Run(logicFn, perConfig)
}

// Unwrap returns the original *gocb.Transactions instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (it *instaTransactions) Unwrap() *gocb.Transactions {
	return it.Transactions
}

// Helper functions

// createTransactions will wrap *gocb.Transactions in to instaTransactions and will return it as Transactions interface
func createTransactions(tracer requestTracer, t *gocb.Transactions) Transactions {
	return &instaTransactions{
		iTracer:      tracer,
		Transactions: t,
	}
}
