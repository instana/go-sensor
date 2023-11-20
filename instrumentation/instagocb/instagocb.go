// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"context"
	"strconv"
	"time"

	"github.com/couchbase/gocb/v2"
	gocbconnstr "github.com/couchbase/gocbcore/v10/connstr"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

const (
	bucketNameSpanTag string = "couchbase.bucket"
	bucketTypeSpanTag string = "couchbase.type"
	operationSpanTag  string = "couchbase.sql"
	errorSpanTag      string = "couchbase.error"

	// keeping this here, for the documentation purpose
	// hostNameSpanTag   string = "couchbase.hostname"
)

// wrapper interface on top of gocb.RequestTracer
type requestTracer interface {
	gocb.RequestTracer
	wrapCluster(cluster Cluster)
}

// Instana tracer
type Tracer struct {
	collector   instana.TracerLogger
	connDetails instana.DbConnDetails
	cluster     Cluster

	// cache for bucket type
	bucketTypeLookup map[string]string
}

// Instana span
type Span struct {
	wrapped       opentracing.Span
	ctx           context.Context
	tracer        *Tracer
	operationType string
	err           error
}

// Request span will create a new span
func (t *Tracer) RequestSpan(parentContext gocb.RequestSpanContext, operationType string) gocb.RequestSpan {
	ctx := context.Background()

	if context, ok := parentContext.(context.Context); ok {
		ctx = context
	}

	s, _ := instana.StartSQLSpan(ctx, t.connDetails, operationType, t.collector)

	return &Span{
		wrapped:       s,
		ctx:           instana.ContextWithSpan(ctx, s),
		tracer:        t,
		operationType: operationType,
	}
}

func (t *Tracer) wrapCluster(cluster Cluster) {
	t.cluster = cluster
}

// Ending a span
func (s *Span) End() {
	if s != nil && s.err != nil {
		s.wrapped.SetTag(errorSpanTag, s.err.Error())
		s.wrapped.SetTag(string(ext.Error), s.err.Error())
		s.wrapped.LogFields(otlog.Object("error", s.err.Error()))
	}

	if s != nil && s.wrapped != nil {
		s.wrapped.Finish()
	}
}

// To get the span context
func (s *Span) Context() gocb.RequestSpanContext {
	if s == nil {
		return context.TODO()
	}
	return s.ctx
}

// Not used; implemented as this one is part of gocb.RequestSpan interface
func (s *Span) AddEvent(name string, timestamp time.Time) {}

// Setting attributes in span
func (s *Span) SetAttribute(key string, value interface{}) {
	if s == nil {
		return
	}

	switch key {
	case bucketNameSpanTag:
		bucketName := value.(string)
		if bucketName == "" {
			break
		}
		if bucketType, ok := s.tracer.bucketTypeLookup[bucketName]; ok {
			s.wrapped.SetTag(bucketTypeSpanTag, bucketType)
			s.wrapped.SetTag(bucketNameSpanTag, bucketName)
			return
		}
		s.wrapped.SetTag(bucketNameSpanTag, bucketName)
		bm := s.tracer.cluster.Buckets()
		bs, err := bm.(*InstanaBucketManager).BucketManager.GetBucket(bucketName, &gocb.GetBucketOptions{})
		if err == nil && bs != nil {
			s.tracer.bucketTypeLookup[bucketName] = string(bs.BucketType)
			s.wrapped.SetTag(bucketTypeSpanTag, string(bs.BucketType))
		}
	}

	s.wrapped.SetTag(key, value)

}

// wrapper for gocb.Connect - it will return an instrumented *gocb.Cluster instance
func Connect(s instana.TracerLogger, connStr string, opts gocb.ClusterOptions) (Cluster, error) {
	// create a new instana tracer
	t := newInstanaTracer(s, connStr)

	cluster, err := gocb.Connect(connStr, opts)

	if err != nil {
		return nil, err
	}

	icluster := &InstanaCluster{
		iTracer: t,
		Cluster: cluster,
	}

	// wrapping the connected cluster in tracer
	t.wrapCluster(icluster)

	return icluster, nil
}

// Getting parent span from current context.
// Users need to pass this parent span to the options (eg : gocb.QueryOptions)
func GetParentSpanFromContext(ctx context.Context) *Span {
	s, ok := instana.SpanFromContext(ctx)

	if !ok {
		return nil
	}

	return &Span{
		wrapped: s,
		ctx:     ctx,
	}
}

// helper functions

// creates a new instana tracer instance
func newInstanaTracer(it instana.TracerLogger, dsn string) requestTracer {
	var raw string

	// parsing connection string
	connSpec, err := gocbconnstr.Parse(dsn)
	if err == nil {
		for i, addr := range connSpec.Addresses {
			if i != 0 {
				raw += ","
			}
			raw += addr.Host
			if addr.Port != -1 {
				raw += ":" + strconv.Itoa(addr.Port)
			}
		}
	} else {
		raw = dsn
	}

	return &Tracer{
		collector: it,
		connDetails: instana.DbConnDetails{
			RawString:    raw,
			DatabaseName: string(instana.CouchbaseSpanType),
		},
		bucketTypeLookup: map[string]string{},
	}

}
