package instagocb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
)

var bucketTypeLookup map[string]string

type requestTracer interface {
	gocb.RequestTracer
	wrapCluster(cluster *gocb.Cluster)
}

type Tracer struct {
	sensor      instana.TracerLogger
	connDetails instana.DbConnDetails
	cluster     *gocb.Cluster
}

type Span struct {
	wrapped         opentracing.Span
	ctx             context.Context
	tracer          *Tracer
	noTracingNeeded bool
}

func (t *Tracer) RequestSpan(parentContext gocb.RequestSpanContext, operationName string) gocb.RequestSpan {
	fmt.Println("Span RequestSpan", operationName)

	if isOperationNameInNotTracedList(operationName) {
		return &Span{
			noTracingNeeded: true,
		}
	}

	ctx := context.Background()

	if context, ok := parentContext.(context.Context); ok {
		ctx = context
	}

	s, _ := instana.StartSQLSpan(ctx, t.connDetails, operationName, t.sensor)
	s.SetTag("couchbase.sql", operationName)

	return &Span{
		wrapped: s,
		ctx:     instana.ContextWithSpan(ctx, s),
		tracer:  t,
	}
}

func (t *Tracer) wrapCluster(cluster *gocb.Cluster) {
	t.cluster = cluster
}

func (s *Span) End() {
	if s != nil && s.wrapped != nil {
		s.wrapped.Finish()
	}
	// fmt.Println("Span end!")

}

func (s *Span) Context() gocb.RequestSpanContext {
	if s == nil {
		return nil
	}
	// fmt.Println("Span Context!")
	return s.ctx
}

func (s *Span) AddEvent(name string, timestamp time.Time) {
	if s == nil {
		return
	}
	// fmt.Println("Span AddEvent!", name, timestamp)
	s.SetAttribute(name, timestamp)
}

func (s *Span) SetAttribute(key string, value interface{}) {
	if s == nil {
		return
	}
	fmt.Println("Span SetAttribute!", key, value)

	if s.noTracingNeeded {
		return
	}

	switch key {
	case "db.name":
		bucketName := value.(string)
		if bucketTypeLookup == nil {
			bucketTypeLookup = make(map[string]string)
		}

		if bucketType, ok := bucketTypeLookup[bucketName]; ok {
			s.wrapped.SetTag("couchbase.type", bucketType)
			s.wrapped.SetTag("couchbase.bucket", bucketName)
			break
		}

		bm := s.tracer.cluster.Buckets()
		bs, _ := bm.GetBucket(bucketName, &gocb.GetBucketOptions{})
		bucketTypeLookup[bucketName] = string(bs.BucketType)
		s.wrapped.SetTag("couchbase.type", bs.BucketType)
		s.wrapped.SetTag("couchbase.bucket", value)

	case "db.statement":
		s.wrapped.SetTag("couchbase.sql", value)
	case "db.operation":
		if str, ok := value.(string); ok {
			s.wrapped.SetTag("couchbase.sql", strings.ToUpper(str))
		} else {
			s.wrapped.SetTag("couchbase.sql", value)
		}
	}

}

// wrapper for gocb.Connect - it will return an instrumented *gocb.Cluster instance
func InstrumentAndConnect(s instana.TracerLogger, connStr string, opts gocb.ClusterOptions) (*gocb.Cluster, error) {
	// create a new instana tracer
	t := newInstanaTracer(s, connStr)
	opts.Tracer = t // adding the instana tracer to couchbase connection options

	cluster, err := gocb.Connect(connStr, opts)

	if err != nil {
		return nil, err
	}

	// wrapping the connected cluster in tracer
	t.wrapCluster(cluster)

	return cluster, nil
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
	}

}

// gocb.RequestTracer traces a lot of operations, we don't need that much tracing happen
// Add any operation here to skip the tracing.
func isOperationNameInNotTracedList(operationName string) bool {
	if strings.HasPrefix(operationName, "CMD") {
		return true
	}

	if operationName == "dispatch_to_server" || operationName == "request_encoding" {
		return true
	}

	// manager_bucket_create_bucket operation can't be traced because we need to call
	// this method to fetch the bucket type internally.
	// If we trace this call, it will create circular calls(dead lock).
	if operationName == "manager_bucket_create_bucket" {
		return true
	}

	return false
}
