// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"github.com/couchbase/gocb/v2"
)

type CouchbaseMap interface {
	Iterator() (map[string]interface{}, error)
	At(id string, valuePtr interface{}) error
	Add(id string, val interface{}) error
	Remove(id string) error
	Exists(id string) (bool, error)
	Size() (int, error)
	Keys() ([]string, error)
	Values() ([]interface{}, error)
	Clear() error

	Unwrap() *gocb.CouchbaseMap
}

type instaCouchbaseMap struct {
	*gocb.CouchbaseMap
	iTracer gocb.RequestTracer

	collection Collection
}

// Iterator returns an iterable for all items in the map.
func (icm *instaCouchbaseMap) Iterator() (map[string]interface{}, error) {
	span := icm.iTracer.RequestSpan(nil, "MAP_ITERATOR")
	span.SetAttribute(bucketNameSpanTag, icm.collection.Bucket().Name())

	result, err := icm.CouchbaseMap.Iterator()

	span.(*Span).err = err

	defer span.End()
	return result, err
}

// At retrieves the item for the given id from the map.
func (icm *instaCouchbaseMap) At(id string, valuePtr interface{}) error {
	span := icm.iTracer.RequestSpan(nil, "MAP_AT")
	span.SetAttribute(bucketNameSpanTag, icm.collection.Bucket().Name())

	err := icm.CouchbaseMap.At(id, valuePtr)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Add adds an item to the map.
func (icm *instaCouchbaseMap) Add(id string, val interface{}) error {
	span := icm.iTracer.RequestSpan(nil, "MAP_ADD")
	span.SetAttribute(bucketNameSpanTag, icm.collection.Bucket().Name())

	err := icm.CouchbaseMap.Add(id, val)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Remove removes an item from the map.
func (icm *instaCouchbaseMap) Remove(id string) error {
	span := icm.iTracer.RequestSpan(nil, "MAP_REMOVE")
	span.SetAttribute(bucketNameSpanTag, icm.collection.Bucket().Name())

	err := icm.CouchbaseMap.Remove(id)

	span.(*Span).err = err

	defer span.End()
	return err
}

// Exists verifies whether or a id exists in the map.
func (icm *instaCouchbaseMap) Exists(id string) (bool, error) {
	span := icm.iTracer.RequestSpan(nil, "MAP_EXISTS")
	span.SetAttribute(bucketNameSpanTag, icm.collection.Bucket().Name())

	res, err := icm.CouchbaseMap.Exists(id)

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Size returns the size of the map.
func (icm *instaCouchbaseMap) Size() (int, error) {
	span := icm.iTracer.RequestSpan(nil, "MAP_SIZE")
	span.SetAttribute(bucketNameSpanTag, icm.collection.Bucket().Name())

	res, err := icm.CouchbaseMap.Size()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Keys returns all of the keys within the map.
func (icm *instaCouchbaseMap) Keys() ([]string, error) {
	span := icm.iTracer.RequestSpan(nil, "MAP_KEYS")
	span.SetAttribute(bucketNameSpanTag, icm.collection.Bucket().Name())

	res, err := icm.CouchbaseMap.Keys()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Values returns all of the values within the map.
func (icm *instaCouchbaseMap) Values() ([]interface{}, error) {
	span := icm.iTracer.RequestSpan(nil, "MAP_VALUES")
	span.SetAttribute(bucketNameSpanTag, icm.collection.Bucket().Name())

	res, err := icm.CouchbaseMap.Values()

	span.(*Span).err = err

	defer span.End()
	return res, err
}

// Clear clears a map, also removing it.
func (icm *instaCouchbaseMap) Clear() error {
	span := icm.iTracer.RequestSpan(nil, "MAP_CLEAR")
	span.SetAttribute(bucketNameSpanTag, icm.collection.Bucket().Name())

	err := icm.CouchbaseMap.Clear()

	span.(*Span).err = err

	defer span.End()
	return err
}

// Unwrap returns the original *gocb.CouchbaseMap instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (icm *instaCouchbaseMap) Unwrap() *gocb.CouchbaseMap {
	return icm.CouchbaseMap
}

// helper functions

func createMap(ic *instaCollection, id string) CouchbaseMap {

	// creating a gocb.CouchbaseMap object.
	m := ic.Collection.Map(id)

	return &instaCouchbaseMap{
		iTracer:      ic.iTracer,
		CouchbaseMap: m,

		collection: ic,
	}
}
