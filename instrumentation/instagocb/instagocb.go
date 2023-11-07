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

type Cluster interface {
	Bucket(bucketName string) Bucket
	Buckets() *gocb.BucketManager
	WaitUntilReady(timeout time.Duration, opts *gocb.WaitUntilReadyOptions) error
	Close(opts *gocb.ClusterCloseOptions) error
	Users() *gocb.UserManager
	AnalyticsIndexes() *gocb.AnalyticsIndexManager
	QueryIndexes() *gocb.QueryIndexManager
	SearchIndexes() *gocb.SearchIndexManager
	EventingFunctions() *gocb.EventingFunctionManager
	Transactions() *gocb.Transactions
}

type InstanaCluster struct {
	t gocb.RequestTracer
	*gocb.Cluster
}

func (ic *InstanaCluster) Bucket(bn string) Bucket {
	b := &InstanaBucket{
		t:      ic.t,
		Bucket: ic.Cluster.Bucket(bn),
	}
	return b
}

type Bucket interface {
	Name() string
	Scope(scopeName string) Scope
	DefaultScope() *gocb.Scope
	Collection(collectionName string) *gocb.Collection
	DefaultCollection() *gocb.Collection
	ViewIndexes() *gocb.ViewIndexManager
	Collections() *gocb.CollectionManager
	WaitUntilReady(timeout time.Duration, opts *gocb.WaitUntilReadyOptions) error
}

type InstanaBucket struct {
	t gocb.RequestTracer
	*gocb.Bucket
}

func (ib *InstanaBucket) Scope(s string) Scope {
	scope := ib.Bucket.Scope(s)

	return &InstanaScope{
		t:     ib.t,
		Scope: scope,
	}
}

type Scope interface {
	Name() string
	BucketName() string
	Collection(collectionName string) *gocb.Collection
	Query(statement string, opts *gocb.QueryOptions) (*gocb.QueryResult, error)
}

type InstanaScope struct {
	t gocb.RequestTracer
	*gocb.Scope
}

func (is *InstanaScope) Query(statement string, opts *gocb.QueryOptions) (*gocb.QueryResult, error) {
	span := is.t.RequestSpan(opts.ParentSpan.Context, "instana-query")
	span.SetAttribute("db.statement", statement)
	span.SetAttribute("db.name", is.BucketName())
	span.SetAttribute("db.couchbase.scope", is.Name())

	res, err := is.Scope.Query(statement, opts)

	span.(*Span).err = err

	span.End()
	fmt.Println(">>> FINISHED SPAN")

	return res, err
}

type requestTracer interface {
	gocb.RequestTracer
	wrapCluster(cluster Cluster)
}

type Tracer struct {
	sensor      instana.TracerLogger
	connDetails instana.DbConnDetails
	cluster     Cluster
}

type Span struct {
	wrapped         opentracing.Span
	ctx             context.Context
	tracer          *Tracer
	noTracingNeeded bool
	operationType   string
	err             error
}

func (t *Tracer) RequestSpan(parentContext gocb.RequestSpanContext, operationType string) gocb.RequestSpan {
	fmt.Println("Span RequestSpan", operationType)

	if isOperationNameInNotTracedList(operationType) {
		return &Span{
			noTracingNeeded: true,
		}
	}

	ctx := context.Background()

	if context, ok := parentContext.(context.Context); ok {
		ctx = context
	}

	s, _ := instana.StartSQLSpan(ctx, t.connDetails, operationType, t.sensor)
	s.SetTag("couchbase.sql", operationType)

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
	// ignore original query span
	if s.operationType == "query" {
		fmt.Println(">>> SHOULD BE HERE ONCE")
		return
	}

	if s.err != nil {
		fmt.Println(">>> SHOULD COLLECT ERROR ONCE:", s.err)
		s.wrapped.SetTag("couchbase.error", s.err)
	}

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
// func InstrumentAndConnect(s instana.TracerLogger, connStr string, opts gocb.ClusterOptions) (*gocb.Cluster, error) {
func InstrumentAndConnect(s instana.TracerLogger, connStr string, opts gocb.ClusterOptions) (Cluster, error) {
	// create a new instana tracer
	t := newInstanaTracer(s, connStr)
	opts.Tracer = t // adding the instana tracer to couchbase connection options

	cluster, err := gocb.Connect(connStr, opts)

	if err != nil {
		return nil, err
	}

	icluster := &InstanaCluster{
		t:       t,
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
