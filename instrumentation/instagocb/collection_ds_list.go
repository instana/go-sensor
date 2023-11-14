// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"github.com/couchbase/gocb/v2"
)

type CouchbaseList interface {
	Iterator() ([]interface{}, error)
	At(index int, valuePtr interface{}) error
	RemoveAt(index int) error
	Append(val interface{}) error
	Prepend(val interface{}) error
	IndexOf(val interface{}) (int, error)
	Size() (int, error)
	Clear() error
}

type InstanaCouchbaseList struct {
	*gocb.CouchbaseList
	iTracer gocb.RequestTracer

	collection Collection
}

// Iterator returns an iterable for all items in the list.
func (icl *InstanaCouchbaseList) Iterator() ([]interface{}, error) {
	span := icl.iTracer.RequestSpan(nil, "LIST_ITERATOR")
	span.SetAttribute(bucketNameSpanTag, icl.collection.Bucket().Name())

	result, err := icl.CouchbaseList.Iterator()

	span.(*Span).err = err

	defer span.End()
	return result, err
}

// At retrieves the value specified at the given index from the list.
func (icl *InstanaCouchbaseList) At(index int, valuePtr interface{}) error {
	span := icl.iTracer.RequestSpan(nil, "LIST_AT")
	span.SetAttribute(bucketNameSpanTag, icl.collection.Bucket().Name())

	err := icl.CouchbaseList.At(index, valuePtr)

	span.(*Span).err = err

	defer span.End()
	return err
}

// RemoveAt removes the value specified at the given index from the list.
func (icl *InstanaCouchbaseList) RemoveAt(index int) error {
	span := icl.iTracer.RequestSpan(nil, "LIST_REMOVE_AT")
	span.SetAttribute(bucketNameSpanTag, icl.collection.Bucket().Name())

	err := icl.CouchbaseList.RemoveAt(index)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Append appends an item to the list.
func (icl *InstanaCouchbaseList) Append(val interface{}) error {
	span := icl.iTracer.RequestSpan(nil, "LIST_APPEND")
	span.SetAttribute(bucketNameSpanTag, icl.collection.Bucket().Name())

	err := icl.CouchbaseList.Append(val)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Prepend prepends an item to the list.
func (icl *InstanaCouchbaseList) Prepend(val interface{}) error {
	span := icl.iTracer.RequestSpan(nil, "LIST_PREPEND")
	span.SetAttribute(bucketNameSpanTag, icl.collection.Bucket().Name())

	err := icl.CouchbaseList.Prepend(val)

	span.(*Span).err = err

	defer span.End()
	return err
}

// IndexOf gets the index of the item in the list.
func (icl *InstanaCouchbaseList) IndexOf(val interface{}) (int, error) {
	span := icl.iTracer.RequestSpan(nil, "LIST_INDEX_OF")
	span.SetAttribute(bucketNameSpanTag, icl.collection.Bucket().Name())

	res, err := icl.CouchbaseList.IndexOf(val)

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Size returns the size of the list.
func (icl *InstanaCouchbaseList) Size() (int, error) {
	span := icl.iTracer.RequestSpan(nil, "LIST_SIZE")
	span.SetAttribute(bucketNameSpanTag, icl.collection.Bucket().Name())

	res, err := icl.CouchbaseList.Size()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Clear clears a list, also removing it.
func (icl *InstanaCouchbaseList) Clear() error {
	span := icl.iTracer.RequestSpan(nil, "LIST_CLEAR")
	span.SetAttribute(bucketNameSpanTag, icl.collection.Bucket().Name())

	err := icl.CouchbaseList.Clear()

	span.(*Span).err = err

	defer span.End()
	return err
}

// helper functions

func createList(ic *InstanaCollection, id string) CouchbaseList {

	// creating a gocb.CouchbaseList object.
	l := ic.Collection.List(id)

	return &InstanaCouchbaseList{
		iTracer:       ic.iTracer,
		CouchbaseList: l,

		collection: ic,
	}
}
