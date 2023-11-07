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

type RequestTracer interface {
	gocb.RequestTracer
	WrapCluster(cluster *gocb.Cluster)
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

// RequestSpan belongs to the Tracer interface.
func (t *Tracer) RequestSpan(parentContext gocb.RequestSpanContext, operationName string) gocb.RequestSpan {
	fmt.Println("Span RequestSpan", parentContext, operationName)

	if operationName == "manager_bucket_create_bucket" {
		return &Span{
			noTracingNeeded: true,
		}
	}

	ctx := context.Background()

	if context, ok := parentContext.(context.Context); ok {
		ctx = context
	}

	s, dbKey := instana.StartSQLSpan(ctx, t.connDetails, operationName, t.sensor)
	s.SetBaggageItem("dbKey", dbKey)

	return &Span{
		wrapped: s,
		ctx:     instana.ContextWithSpan(ctx, s),
		tracer:  t,
	}
}

func (t *Tracer) WrapCluster(cluster *gocb.Cluster) {
	t.cluster = cluster
}

func (s *Span) End() {
	if s != nil && s.wrapped != nil {
		s.wrapped.Finish()
	}
	fmt.Println("Span end!")

}

func (s *Span) Context() gocb.RequestSpanContext {
	if s == nil {
		return nil
	}
	fmt.Println("Span Context!")
	return s.ctx
}

func (s *Span) AddEvent(name string, timestamp time.Time) {
	if s == nil {
		return
	}
	fmt.Println("Span AddEvent!", name, timestamp)
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
		bm := s.tracer.cluster.Buckets()
		bs, _ := bm.GetBucket(value.(string), &gocb.GetBucketOptions{})
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

func NewTracer(s instana.TracerLogger, dsn string) RequestTracer {
	return &Tracer{
		sensor: s,
		connDetails: instana.DbConnDetails{
			RawString:    dsn,
			DatabaseName: string(instana.CouchbaseSpanType),
		},
	}
}

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
