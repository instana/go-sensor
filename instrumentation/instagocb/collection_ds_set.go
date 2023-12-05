// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"github.com/couchbase/gocb/v2"
)

type CouchbaseSet interface {
	Iterator() ([]interface{}, error)
	Add(val interface{}) error
	Remove(val string) error
	Values() ([]interface{}, error)
	Contains(val string) (bool, error)
	Size() (int, error)
	Clear() error

	Unwrap() *gocb.CouchbaseSet
}

type instaCouchbaseSet struct {
	*gocb.CouchbaseSet
	iTracer gocb.RequestTracer

	collection Collection
}

// Iterator returns an iterable for all items in the set.
func (ics *instaCouchbaseSet) Iterator() ([]interface{}, error) {
	span := ics.iTracer.RequestSpan(nil, "SET_ITERATOR")
	span.SetAttribute(bucketNameSpanTag, ics.collection.Bucket().Name())

	res, err := ics.CouchbaseSet.Iterator()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Add adds a value to the set.
func (ics *instaCouchbaseSet) Add(val interface{}) error {
	span := ics.iTracer.RequestSpan(nil, "SET_ADD")
	span.SetAttribute(bucketNameSpanTag, ics.collection.Bucket().Name())

	err := ics.CouchbaseSet.Add(val)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Remove removes an value from the set.
func (ics *instaCouchbaseSet) Remove(val string) error {
	span := ics.iTracer.RequestSpan(nil, "SET_REMOVE")
	span.SetAttribute(bucketNameSpanTag, ics.collection.Bucket().Name())

	err := ics.CouchbaseSet.Remove(val)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Values returns all of the values within the set.
func (ics *instaCouchbaseSet) Values() ([]interface{}, error) {
	span := ics.iTracer.RequestSpan(nil, "SET_VALUES")
	span.SetAttribute(bucketNameSpanTag, ics.collection.Bucket().Name())

	res, err := ics.CouchbaseSet.Values()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Contains verifies whether or not a value exists within the set.
func (ics *instaCouchbaseSet) Contains(val string) (bool, error) {
	span := ics.iTracer.RequestSpan(nil, "SET_CONTAINS")
	span.SetAttribute(bucketNameSpanTag, ics.collection.Bucket().Name())

	res, err := ics.CouchbaseSet.Contains(val)

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Size returns the size of the set
func (ics *instaCouchbaseSet) Size() (int, error) {
	span := ics.iTracer.RequestSpan(nil, "SET_SIZE")
	span.SetAttribute(bucketNameSpanTag, ics.collection.Bucket().Name())

	res, err := ics.CouchbaseSet.Size()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Clear clears a set, also removing it.
func (ics *instaCouchbaseSet) Clear() error {
	span := ics.iTracer.RequestSpan(nil, "SET_CLEAR")
	span.SetAttribute(bucketNameSpanTag, ics.collection.Bucket().Name())

	err := ics.CouchbaseSet.Clear()

	span.(*Span).err = err

	defer span.End()
	return err
}

// Unwrap returns the original *gocb.CouchbaseSet instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (ics *instaCouchbaseSet) Unwrap() *gocb.CouchbaseSet {
	return ics.CouchbaseSet
}

// helper functions

func createSet(ic *instaCollection, id string) CouchbaseSet {

	// creating a gocb.CouchbaseSet object.
	s := ic.Collection.Set(id)

	return &instaCouchbaseSet{
		iTracer:      ic.iTracer,
		CouchbaseSet: s,

		collection: ic,
	}
}
