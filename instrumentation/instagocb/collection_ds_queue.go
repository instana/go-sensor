// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"github.com/couchbase/gocb/v2"
)

type CouchbaseQueue interface {
	Iterator() ([]interface{}, error)
	Push(val interface{}) error
	Pop(valuePtr interface{}) error
	Size() (int, error)
	Clear() error

	Unwrap() *gocb.CouchbaseQueue
}

type instaCouchbaseQueue struct {
	*gocb.CouchbaseQueue
	iTracer gocb.RequestTracer

	collection Collection
}

// Iterator returns an iterable for all items in the queue.
func (icq *instaCouchbaseQueue) Iterator() ([]interface{}, error) {
	span := icq.iTracer.RequestSpan(nil, "QUEUE_ITERATOR")
	span.SetAttribute(bucketNameSpanTag, icq.collection.Bucket().Name())

	res, err := icq.CouchbaseQueue.Iterator()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Push pushes a value onto the queue.
func (icq *instaCouchbaseQueue) Push(val interface{}) error {
	span := icq.iTracer.RequestSpan(nil, "QUEUE_PUSH")
	span.SetAttribute(bucketNameSpanTag, icq.collection.Bucket().Name())

	err := icq.CouchbaseQueue.Push(val)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Pop pops an items off of the queue.
func (icq *instaCouchbaseQueue) Pop(valuePtr interface{}) error {
	span := icq.iTracer.RequestSpan(nil, "QUEUE_POP")
	span.SetAttribute(bucketNameSpanTag, icq.collection.Bucket().Name())

	err := icq.CouchbaseQueue.Pop(valuePtr)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Size returns the size of the queue.
func (icq *instaCouchbaseQueue) Size() (int, error) {
	span := icq.iTracer.RequestSpan(nil, "QUEUE_SIZE")
	span.SetAttribute(bucketNameSpanTag, icq.collection.Bucket().Name())

	res, err := icq.CouchbaseQueue.Size()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Clear clears a queue, also removing it.
func (icq *instaCouchbaseQueue) Clear() error {
	span := icq.iTracer.RequestSpan(nil, "QUEUE_CLEAR")
	span.SetAttribute(bucketNameSpanTag, icq.collection.Bucket().Name())

	err := icq.CouchbaseQueue.Clear()

	span.(*Span).err = err

	defer span.End()
	return err
}

// Unwrap returns the original *gocb.CouchbaseQueue instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (icq *instaCouchbaseQueue) Unwrap() *gocb.CouchbaseQueue {
	return icq.CouchbaseQueue
}

// helper functions

func createQueue(ic *instaCollection, id string) CouchbaseQueue {

	// creating a gocb.CouchbaseQueue object.
	q := ic.Collection.Queue(id)

	return &instaCouchbaseQueue{
		iTracer:        ic.iTracer,
		CouchbaseQueue: q,

		collection: ic,
	}
}
