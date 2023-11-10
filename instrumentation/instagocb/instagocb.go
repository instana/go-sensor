// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"context"
	"time"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

const (
	hostNameSpanTag   string = "couchbase.hostname"
	bucketNameSpanTag string = "couchbase.bucket"
	bucketTypeSpanTag string = "couchbase.type"
	operationSpanTag  string = "couchbase.sql"
	errorSpanTag      string = "couchbase.error"
)

type requestTracer interface {
	gocb.RequestTracer
	wrapCluster(cluster Cluster)
}

type Tracer struct {
	sensor      instana.TracerLogger
	connDetails instana.DbConnDetails
	cluster     Cluster

	// cache for bucket type
	bucketTypeLookup map[string]string
}

type Span struct {
	wrapped       opentracing.Span
	ctx           context.Context
	tracer        *Tracer
	operationType string
	err           error
}

func (t *Tracer) RequestSpan(parentContext gocb.RequestSpanContext, operationType string) gocb.RequestSpan {
	// fmt.Println("Span RequestSpan", operationType)

	ctx := context.Background()

	if context, ok := parentContext.(context.Context); ok {
		ctx = context
	}

	s, _ := instana.StartSQLSpan(ctx, t.connDetails, operationType, t.sensor)

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

func (s *Span) Context() gocb.RequestSpanContext {
	if s == nil {
		return nil
	}
	return s.ctx
}

func (s *Span) AddEvent(name string, timestamp time.Time) {
	if s == nil {
		return
	}
	s.SetAttribute(name, timestamp)
}

func (s *Span) SetAttribute(key string, value interface{}) {
	if s == nil {
		return
	}

	switch key {
	case bucketNameSpanTag:
		bucketName := value.(string)
		if bucketType, ok := s.tracer.bucketTypeLookup[bucketName]; ok {
			s.wrapped.SetTag(bucketTypeSpanTag, bucketType)
			s.wrapped.SetTag(bucketNameSpanTag, bucketName)
			return
		}
		bm := s.tracer.cluster.Buckets()
		bs, _ := bm.GetBucket(bucketName, &gocb.GetBucketOptions{})
		s.tracer.bucketTypeLookup[bucketName] = string(bs.BucketType)
		s.wrapped.SetTag(bucketTypeSpanTag, bs.BucketType)
		s.wrapped.SetTag(bucketNameSpanTag, bucketName)
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

// Getting parent span from current context - users need to pass this parent span to the options (eg : gocb.QueryOptions)
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
func newInstanaTracer(s instana.TracerLogger, dsn string) requestTracer {
	return &Tracer{
		sensor: s,
		connDetails: instana.DbConnDetails{
			RawString:    dsn,
			DatabaseName: string(instana.CouchbaseSpanType),
		},
		bucketTypeLookup: map[string]string{},
	}

}
