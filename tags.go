// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"strings"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	batchSizeTag       = "batch_size"
	suppressTracingTag = "suppress_tracing"
	syntheticCallTag   = "synthetic_call"
)

// Tags is the interface that applies new tags to ot.Tags map
type Tags interface {
	Apply(ot.Tags)
}

// TagsFunc implements Tags interface
type TagsFunc func(ot.Tags)

func (f TagsFunc) Apply(t ot.Tags) {
	f(t)
}

// DBTagsFunc is a function type that takes a DbConnDetails struct as input and returns Tags interface.
// It can be used to generate or retrieve a set of tags associated with a specific database connection.
type DBTagsFunc func(c DbConnDetails) Tags

var tagsFuncMap = map[database]DBTagsFunc{
	postgres:    withPostgresTags,
	mysql:       withMySQLTags,
	redis:       withRedisTags,
	couchbase:   withCouchbaseTags,
	cosmos:      withCosmosTags,
	sql_generic: withGenericSQLTags,
}

var withPostgresTags DBTagsFunc = func(c DbConnDetails) Tags {
	return TagsFunc(func(tags ot.Tags) {

		tags["pg.user"] = c.User
		tags["pg.host"] = c.Host

		if c.Schema != "" {
			tags["pg.db"] = c.Schema
		} else {
			tags["pg.db"] = c.RawString
		}

		if c.Port != "" {
			tags["pg.port"] = c.Port
		}

	})
}

var withMySQLTags DBTagsFunc = func(c DbConnDetails) Tags {
	return TagsFunc(func(tags ot.Tags) {

		tags["mysql.user"] = c.User
		tags["mysql.host"] = c.Host

		if c.Schema != "" {
			tags["mysql.db"] = c.Schema
		} else {
			tags["mysql.db"] = c.RawString
		}

		if c.Port != "" {
			tags["mysql.port"] = c.Port
		}

	})
}

var withRedisTags DBTagsFunc = func(c DbConnDetails) Tags {
	return TagsFunc(func(tags ot.Tags) {

		if c.Error != nil {
			tags["redis.error"] = c.Error.Error()
		}

		connection := c.Host + ":" + c.Port

		if c.Host == "" || c.Port == "" {
			i := strings.LastIndex(c.RawString, "@")
			connection = c.RawString[i+1:]
		}

		tags["redis.connection"] = connection

	})
}

var withCouchbaseTags DBTagsFunc = func(c DbConnDetails) Tags {
	return TagsFunc(func(tags ot.Tags) {
		tags["couchbase.hostname"] = c.RawString
	})
}

var withCosmosTags DBTagsFunc = func(c DbConnDetails) Tags {
	return TagsFunc(func(tags ot.Tags) {})
}

var withGenericSQLTags DBTagsFunc = func(c DbConnDetails) Tags {
	return TagsFunc(func(tags ot.Tags) {

		tags[string(ext.DBType)] = "sql"
		tags[string(ext.PeerAddress)] = c.RawString

		if c.Schema != "" {
			tags[string(ext.DBInstance)] = c.Schema
		} else {
			tags[string(ext.DBInstance)] = c.RawString
		}

		if c.Host != "" {
			tags[string(ext.PeerHostname)] = c.Host
		}

		if c.Port != "" {
			tags[string(ext.PeerPort)] = c.Port
		}
	})
}

// BatchSize returns an opentracing.Tag to mark the span as a batched span representing
// similar span categories. An example of such span would be batch writes to a queue,
// a database, etc. If the batch size less than 2, then this option has no effect
func BatchSize(n int) ot.Tag {
	return ot.Tag{Key: batchSizeTag, Value: n}
}

// SuppressTracing returns an opentracing.Tag to mark the span and any of its child spans
// as not to be sent to the agent
func SuppressTracing() ot.Tag {
	return ot.Tag{Key: suppressTracingTag, Value: true}
}

func syntheticCall() ot.Tag {
	return ot.Tag{Key: syntheticCallTag, Value: true}
}
