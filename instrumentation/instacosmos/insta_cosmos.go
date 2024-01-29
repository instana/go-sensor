// (c) Copyright IBM Corp. 2024

//go:build go1.18
// +build go1.18

// Package instacosmos provides instrumentation for the azcosmos package in azure-sdk-for-go
package instacosmos

import (
	"context"
	"fmt"

	instana "github.com/instana/go-sensor"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	otlog "github.com/opentracing/opentracing-go/log"
)

// instrumenting fields
const (
	dataURL          = "cosmos.con"
	dataDB           = "cosmos.db"
	dataCommand      = "cosmos.cmd"
	dataObj          = "cosmos.obj"
	dataPartitionKey = "cosmos.pk"
	dataType         = "cosmos.type"
	dataReturnCode   = "cosmos.rt"
	operationType    = "cosmos.sql"
	dataError        = "error"
)

// events
const (
	errorEvent = "error"
)

// newTracer creates a Tracer with the help of azure tracing package
// it provides the definition for the new span function which creates new span instances
// ctx - The context for creating new tracer
// collector - instana go collector
// connDetails - db connection details for instana
func newTracer(ctx context.Context,
	collector instana.TracerLogger,
	connDetails instana.DbConnDetails) tracing.Tracer {

	t := tracing.NewTracer(func(ctx context.Context, spanName string, options *tracing.SpanOptions) (context.Context, tracing.Span) {
		var oType string
		for _, attr := range options.Attributes {
			if attr.Key == dataCommand {
				oType = attr.Value.(string)
				break
			}
		}
		cosmosSpan, _ := instana.StartSQLSpan(ctx, connDetails, oType, collector)
		return ctx, tracing.NewSpan(tracing.SpanImpl{
			End: func() {
				cosmosSpan.Finish()
			},
			SetAttributes: func(a ...tracing.Attribute) {
				for _, i := range a {
					cosmosSpan.SetTag(i.Key, i.Value)
				}
			},
			AddEvent: func(s string, a ...tracing.Attribute) {
				switch s {
				case errorEvent:
					for _, i := range a {
						cosmosSpan.LogFields(otlog.Object(i.Key, i.Value))
					}
				}
			},
			SetStatus: func(ss tracing.SpanStatus, s string) {
				cosmosSpan.SetTag(s, fmt.Sprintf("%d", ss))
			},
		})

	}, &tracing.TracerOptions{})

	return t
}
