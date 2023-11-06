package instagocb

import (
	"context"
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
)

type Tracer struct {
	sensor      instana.TracerLogger
	connDetails instana.DbConnDetails
}

type Span struct {
	wrapped opentracing.Span
	ctx     context.Context
}

// RequestSpan belongs to the Tracer interface.
func (t *Tracer) RequestSpan(parentContext gocb.RequestSpanContext, operationName string) gocb.RequestSpan {
	fmt.Println("Span RequestSpan", parentContext, operationName)

	ctx := context.Background()

	if context, ok := parentContext.(context.Context); ok {
		ctx = context
	}

	s, dbKey := instana.StartSQLSpan(ctx, t.connDetails, operationName, t.sensor)
	s.SetBaggageItem("dbKey", dbKey)

	return &Span{
		wrapped: s,
		ctx:     instana.ContextWithSpan(ctx, s),
	}
}

func (s *Span) End() {
	fmt.Println("Span end!")
	if s.wrapped != nil {
		s.wrapped.Finish()
	}
}

func (s *Span) Context() gocb.RequestSpanContext {
	fmt.Println("Span Context!")
	return s.ctx
}

func (s *Span) AddEvent(name string, timestamp time.Time) {
	fmt.Println("Span AddEvent!", name, timestamp)
	s.SetAttribute(name, timestamp)
}

func (s *Span) SetAttribute(key string, value interface{}) {
	fmt.Println("Span SetAttribute!", key, value)
	s.wrapped.SetTag(key, value)
}

func NewTracer(s instana.TracerLogger, dsn string) gocb.RequestTracer {
	return &Tracer{
		sensor:      s,
		connDetails: instana.ParseDBConnDetails(dsn),
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
