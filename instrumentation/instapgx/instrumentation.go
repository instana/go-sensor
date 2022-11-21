// (c) Copyright IBM Corp. 2022

package instapgx

import (
	"context"
	"reflect"

	otlog "github.com/opentracing/opentracing-go/log"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	instana "github.com/instana/go-sensor"
	"github.com/jackc/pgx/v4"
)

// Connect establishes a connection with a PostgreSQL server using connection string
// and returns instrumented connection.
func Connect(ctx context.Context, sensor *instana.Sensor, connString string) (*Conn, error) {
	connConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	return ConnectConfig(ctx, sensor, connConfig)
}

// ConnectConfig establishes a connection with a PostgreSQL server using configuration struct
// and returns instrumented connection.
func ConnectConfig(ctx context.Context, sensor *instana.Sensor, connConfig *pgx.ConnConfig) (*Conn, error) {
	c, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, err
	}

	return &Conn{
		c,
		sensor,
		connConfig,
	}, nil
}

func contextWithChildSpan(ctx context.Context, sql string, config *pgx.ConnConfig, sensor *instana.Sensor) (context.Context, ot.Span) {
	var spanOptions []ot.StartSpanOption
	if parent, ok := instana.SpanFromContext(ctx); ok {
		spanOptions = append(spanOptions, ot.ChildOf(parent.Context()))
	}

	host := config.Host
	port := config.Port
	user := config.User
	db := config.Database

	span := sensor.Tracer().StartSpan(string(instana.PostgreSQLSpanType), spanOptions...)
	span.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCClientEnum))
	span.SetTag("pg.db", db)
	span.SetTag("pg.user", user)
	span.SetTag("pg.stmt", sql)
	span.SetTag("pg.host", host)
	span.SetTag("pg.port", port)

	return instana.ContextWithSpan(ctx, span), span
}

func recordAnError(span ot.Span, err error) {
	span.SetTag("pg.error", err.Error())
	span.LogFields(otlog.Object("error", err.Error()))
}

func appendBatchDetails(b *pgx.Batch, sql string) string {
	v := reflect.ValueOf(*b)
	items := v.FieldByName("items")

	if items.Kind() == reflect.Slice {
		for i := 0; i < items.Len(); i++ {
			if detailedBatchModeMaxEntries == i-1 {
				sql += "\n-- ..."
				break
			}
			s := items.Index(i).Elem().FieldByName("query").String()
			sql += "\n" + s + ";"
		}
	}
	return sql
}
