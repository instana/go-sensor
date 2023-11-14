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
}

type InstanaCouchbaseQueue struct {
	*gocb.CouchbaseQueue
	iTracer gocb.RequestTracer

	collection Collection
}

// Iterator returns an iterable for all items in the queue.
func (icq *InstanaCouchbaseQueue) Iterator() ([]interface{}, error) {
	span := icq.iTracer.RequestSpan(nil, "QUEUE_ITERATOR")
	span.SetAttribute(bucketNameSpanTag, icq.collection.Bucket().Name())

	res, err := icq.CouchbaseQueue.Iterator()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Push pushes a value onto the queue.
func (icq *InstanaCouchbaseQueue) Push(val interface{}) error {
	span := icq.iTracer.RequestSpan(nil, "QUEUE_PUSH")
	span.SetAttribute(bucketNameSpanTag, icq.collection.Bucket().Name())

	err := icq.CouchbaseQueue.Push(val)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Pop pops an items off of the queue.
func (icq *InstanaCouchbaseQueue) Pop(valuePtr interface{}) error {
	span := icq.iTracer.RequestSpan(nil, "QUEUE_POP")
	span.SetAttribute(bucketNameSpanTag, icq.collection.Bucket().Name())

	err := icq.CouchbaseQueue.Pop(valuePtr)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Size returns the size of the queue.
func (icq *InstanaCouchbaseQueue) Size() (int, error) {
	span := icq.iTracer.RequestSpan(nil, "QUEUE_SIZE")
	span.SetAttribute(bucketNameSpanTag, icq.collection.Bucket().Name())

	res, err := icq.CouchbaseQueue.Size()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Clear clears a queue, also removing it.
func (icq *InstanaCouchbaseQueue) Clear() error {
	span := icq.iTracer.RequestSpan(nil, "QUEUE_CLEAR")
	span.SetAttribute(bucketNameSpanTag, icq.collection.Bucket().Name())

	err := icq.CouchbaseQueue.Clear()

	span.(*Span).err = err

	defer span.End()
	return err
}

// helper functions

func createQueue(ic *InstanaCollection, id string) CouchbaseQueue {

	// creating a gocb.CouchbaseQueue object.
	q := ic.Collection.Queue(id)

	return &InstanaCouchbaseQueue{
		iTracer:        ic.iTracer,
		CouchbaseQueue: q,

		collection: ic,
	}
}
