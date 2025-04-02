// (c) Copyright IBM Corp. 2025

package instamongo

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var unmarshalReg *bson.Registry

func init() {
	unmarshalReg = bson.NewRegistry()
	unmarshalReg.RegisterTypeMapEntry(bson.TypeEmbeddedDocument, reflect.TypeOf(bson.M{}))
}

// Connect creates and instruments a new mongo.Client
//
// This is a wrapper method for mongo.Connect(), see https://pkg.go.dev/go.mongodb.org/mongo-driver/v2/mongo#Connect for details on
// the original method.
func Connect(sensor instana.TracerLogger, opts ...*options.ClientOptions) (*mongo.Client, error) {
	return mongo.Connect(addInstrumentedCommandMonitor(opts, sensor)...)
}

func addInstrumentedCommandMonitor(opts []*options.ClientOptions, sensor instana.TracerLogger) []*options.ClientOptions {
	// search for the last client options containing a CommandMonitor and wrap it to preserve
	for i := len(opts) - 1; i >= 0; i-- {
		if opts[i] != nil && opts[i].Monitor != nil {
			opts[i].Monitor = WrapCommandMonitor(opts[i].Monitor, sensor)

			return opts
		}
	}

	// if there is no CommandMonitor specified, add one
	return append(opts, &options.ClientOptions{
		Monitor: NewCommandMonitor(sensor),
	})
}

type wrappedCommandMonitor struct {
	mon    *event.CommandMonitor
	sensor instana.TracerLogger
	spans  *spanRegistry
}

// NewCommandMonitor creates a new event.CommandMonitor that instruments a mongo.Client with Instana.
func NewCommandMonitor(sensor instana.TracerLogger) *event.CommandMonitor {
	return WrapCommandMonitor(nil, sensor)
}

// WrapCommandMonitor wraps an existing event.CommandMonitor to instrument a mongo.Client with Instana
func WrapCommandMonitor(mon *event.CommandMonitor, sensor instana.TracerLogger) *event.CommandMonitor {
	wrapper := &wrappedCommandMonitor{
		mon:    mon,
		sensor: sensor,
		spans:  newSpanRegistry(),
	}

	return &event.CommandMonitor{
		Started:   wrapper.Started,
		Succeeded: wrapper.Succeeded,
		Failed:    wrapper.Failed,
	}
}

// Started traces command start initiating a new span. This span is finalized whenever either
// Succeeded() or Failed() method is called with an event containing the same RequestID.
func (m *wrappedCommandMonitor) Started(ctx context.Context, evt *event.CommandStartedEvent) {
	if m.mon != nil && m.mon.Started != nil {
		defer m.mon.Started(ctx, evt)
	}

	ns := evt.DatabaseName
	if collection, ok := evt.Command.Lookup(evt.CommandName).StringValueOK(); ok {
		ns += "." + collection
	}

	// an exit span will be created without a parent span
	// and forwarded if user chose to opt in
	opts := []opentracing.StartSpanOption{
		ext.SpanKindRPCClient,
		m.extractSpanTags(evt),
	}

	parent, ok := instana.SpanFromContext(ctx)
	if ok {
		opts = append(opts, opentracing.ChildOf(parent.Context()))
	}

	sp := m.sensor.Tracer().StartSpan("mongo", opts...)

	m.spans.Add(evt.RequestID, sp)
}

// Succeeded finalizes the command span started by Started()
func (m *wrappedCommandMonitor) Succeeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	if m.mon != nil && m.mon.Succeeded != nil {
		m.mon.Succeeded(ctx, evt)
	}

	sp, ok := m.spans.Remove(evt.RequestID)
	if !ok {
		return
	}

	sp.Finish()
}

// Failed finalizes the command span started by Started() and logs the failure reason
func (m *wrappedCommandMonitor) Failed(ctx context.Context, evt *event.CommandFailedEvent) {
	if m.mon != nil && m.mon.Failed != nil {
		defer m.mon.Failed(ctx, evt)
	}

	sp, ok := m.spans.Remove(evt.RequestID)
	if !ok {
		return
	}
	defer sp.Finish()

	sp.SetTag("mongo.error", evt.Failure.Error())
	sp.LogFields(otlog.Object("error", evt.Failure.Error()))
}

func (m *wrappedCommandMonitor) extractSpanTags(evt *event.CommandStartedEvent) opentracing.Tags {
	ns := evt.DatabaseName
	if collection, ok := evt.Command.Lookup(evt.CommandName).StringValueOK(); ok {
		ns += "." + collection
	}

	tags := opentracing.Tags{
		"mongo.service":   extractAddress(evt.ConnectionID),
		"mongo.namespace": ns,
		"mongo.command":   evt.CommandName,
	}

	validateAndSetTags := func(doc bson.RawValue, tagKey, tagStr string) {
		if err := doc.Validate(); err == nil {
			if data, err := bsonToJSON(doc); err != nil {
				m.sensor.Logger().Warn("failed to marshal mongodb ", evt.CommandName, " ", tagStr, " to json: ", err)
			} else {
				key := "mongo." + tagKey
				tags[key] = string(data)
			}
		}
	}

	validateAndSetTags(evt.Command.Lookup("query"), "query", "query")
	validateAndSetTags(evt.Command.Lookup("filter"), "filter", "filter")
	validateAndSetTags(extractCommandDocument(evt.CommandName, evt.Command), "json", "document")

	return tags
}

// extractAddress extracts the MongoDB server address (either host:port or a path to UNIX socket) from connection ID by
// truncating the optional connection number at the end of the value.
//
// See newConnection() in https://github.com/mongodb/mongo-go-driver/blob/master/x/mongo/driver/topology/connection.go for details
func extractAddress(connID string) string {
	for i := len(connID) - 1; i >= 0; i-- {
		switch connID[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		case '-':
			return connID[0:i]
		default:
			return connID
		}
	}

	return connID
}

func extractCommandDocument(cmdName string, cmdBody bson.Raw) bson.RawValue {
	var v bson.RawValue

	switch cmdName {
	case "insert":
		v = cmdBody.Lookup("documents")
	case "update":
		v = cmdBody.Lookup("updates")
	case "delete":
		v = cmdBody.Lookup("deletes")
	case "aggregate":
		v = cmdBody.Lookup("pipeline")
	case "mapReduce":
		doc := make(map[string]string, 2)

		findAndApplyAttr := func(key string) {
			if val := cmdBody.Lookup(key); val.Validate() == nil {
				if s, ok := stringOrJavaScriptOK(val); ok {
					doc[key] = s
				}
			}
		}

		findAndApplyAttr("map")
		findAndApplyAttr("reduce")

		if typ, data, err := bson.MarshalValue(doc); err == nil {
			v = bson.RawValue{Type: typ, Value: data}
		}
	}

	return v
}

func bsonToJSON(data bson.RawValue) ([]byte, error) {
	var v interface{}

	if err := data.UnmarshalWithRegistry(unmarshalReg, &v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bson value: %s", err)
	}

	return json.Marshal(v)
}

func stringOrJavaScriptOK(v bson.RawValue) (string, bool) {
	if s, ok := v.JavaScriptOK(); ok {
		return s, true
	}

	return v.StringValueOK()
}
