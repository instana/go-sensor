// (c) Copyright IBM Corp. 2024

// Package instapgx provides Instana instrumentation for pgx/v5 package.
package instapgx

import (
	"context"
	"fmt"
	"strconv"

	instana "github.com/instana/go-sensor"
	"github.com/jackc/pgx/v5"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

type pgxTracer struct {
	col       instana.TracerLogger
	dbDetails dbConfig
}

type dbConfig struct {
	Host     string
	Port     string
	User     string
	Database string
}

// InstanaTracer returns Instana tracer which can be used for instrumenting pgx/v5 	database calls.
func InstanaTracer(cfg *pgx.ConnConfig, collector instana.TracerLogger) pgx.QueryTracer {
	if cfg == nil {
		collector.Logger().Error("cfg is nil. Check your database URL")
		return nil
	}

	tr := pgxTracer{
		col: collector,
		dbDetails: dbConfig{
			Host:     cfg.Host,
			Port:     strconv.Itoa(int(cfg.Port)),
			User:     cfg.User,
			Database: cfg.Database,
		},
	}

	return tr
}

func (tr pgxTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	var spanOptions []ot.StartSpanOption
	if parent, ok := instana.SpanFromContext(ctx); ok {
		spanOptions = append(spanOptions, ot.ChildOf(parent.Context()))
	}

	span := tr.col.StartSpan(string(instana.PostgreSQLSpanType), spanOptions...)

	setSpanTags(span, tr.dbDetails, data.SQL)

	ctx = instana.ContextWithSpan(ctx, span)
	return ctx
}

func (tr pgxTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	extractAndFinishSpan(ctx, data.Err)
}

func (tr pgxTracer) TraceBatchStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceBatchStartData) context.Context {
	sqlStmt := "BEGIN"
	var spanOptions []ot.StartSpanOption
	if parent, ok := instana.SpanFromContext(ctx); ok {
		spanOptions = append(spanOptions, ot.ChildOf(parent.Context()))
	}

	span := tr.col.StartSpan(string(instana.PostgreSQLSpanType), spanOptions...)
	defer span.Finish()

	setSpanTags(span, tr.dbDetails, sqlStmt)

	ctx = instana.ContextWithSpan(ctx, span)
	return ctx
}

func (tr pgxTracer) TraceBatchQuery(ctx context.Context, _ *pgx.Conn, data pgx.TraceBatchQueryData) {
	var spanOptions []ot.StartSpanOption
	if parent, ok := instana.SpanFromContext(ctx); ok {
		spanOptions = append(spanOptions, ot.ChildOf(parent.Context()))
	}

	span := tr.col.StartSpan(string(instana.PostgreSQLSpanType), spanOptions...)
	defer span.Finish()

	setSpanTags(span, tr.dbDetails, data.SQL)
}

func (tr pgxTracer) TraceBatchEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceBatchEndData) {
	var span ot.Span
	var ok bool

	if span, ok = instana.SpanFromContext(ctx); !ok {
		return
	}
	defer span.Finish()

	tr.addChildSpanWithResult(span, data.Err)
}

func (tr pgxTracer) addChildSpanWithResult(span ot.Span, err error) {
	var childSpan ot.Span
	var childSpanOpts []ot.StartSpanOption

	childSpanOpts = append(childSpanOpts, ot.ChildOf(span.Context()))
	childSpan = tr.col.StartSpan(string(instana.PostgreSQLSpanType), childSpanOpts...)
	defer childSpan.Finish()

	if err != nil {
		recordError(span, err)
		setSpanTags(childSpan, tr.dbDetails, "ROLLBACK")
	} else {
		setSpanTags(childSpan, tr.dbDetails, "COMMIT")
	}
}

func (tr pgxTracer) TraceCopyFromStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceCopyFromStartData) context.Context {
	sqlStmt := "PGCOPY"
	var spanOptions []ot.StartSpanOption
	if parent, ok := instana.SpanFromContext(ctx); ok {
		spanOptions = append(spanOptions, ot.ChildOf(parent.Context()))
	}

	span := tr.col.StartSpan(string(instana.PostgreSQLSpanType), spanOptions...)

	setSpanTags(span, tr.dbDetails, sqlStmt)

	ctx = instana.ContextWithSpan(ctx, span)
	return ctx
}

func (tr pgxTracer) TraceCopyFromEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceCopyFromEndData) {
	extractAndFinishSpan(ctx, data.Err)
}

func (tr pgxTracer) TracePrepareStart(ctx context.Context, _ *pgx.Conn, data pgx.TracePrepareStartData) context.Context {
	sqlStmt := fmt.Sprintf("PREPARE %s AS %s", data.Name, data.SQL)
	var spanOptions []ot.StartSpanOption
	if parent, ok := instana.SpanFromContext(ctx); ok {
		spanOptions = append(spanOptions, ot.ChildOf(parent.Context()))
	}

	span := tr.col.StartSpan(string(instana.PostgreSQLSpanType), spanOptions...)

	setSpanTags(span, tr.dbDetails, sqlStmt)

	ctx = instana.ContextWithSpan(ctx, span)
	return ctx
}

func (tr pgxTracer) TracePrepareEnd(ctx context.Context, _ *pgx.Conn, data pgx.TracePrepareEndData) {
	extractAndFinishSpan(ctx, data.Err)
}

func extractAndFinishSpan(ctx context.Context, err error) {
	var span ot.Span
	var ok bool

	if span, ok = instana.SpanFromContext(ctx); !ok {
		return
	}
	defer span.Finish()

	if err != nil {
		recordError(span, err)
	}
}

func setSpanTags(span ot.Span, dbDetails dbConfig, sqlStmt string) {
	span.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCClientEnum))
	span.SetTag("pg.db", dbDetails.Database)
	span.SetTag("pg.user", dbDetails.User)
	span.SetTag("pg.host", dbDetails.Host)
	span.SetTag("pg.port", dbDetails.Port)
	span.SetTag("pg.stmt", sqlStmt)
}

func recordError(span ot.Span, err error) {
	span.SetTag("pg.error", err.Error())
	span.LogFields(otlog.Object("error", err.Error()))
}
