// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instamongo

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/event"
)

var unmarshalReg *bsoncodec.Registry

func init() {
	rb := bson.NewRegistryBuilder()
	rb.RegisterTypeMapEntry(bsontype.EmbeddedDocument, reflect.TypeOf(bson.M{}))

	unmarshalReg = rb.Build()
}

type commandMonitor struct {
	sensor *instana.Sensor
	spans  *spanRegistry
}

// NewCommandMonitor creates a new event.CommandMonitor to be used in mongo.ClientOptions.
// This monitor instruments mongo.Client with Instana tracing any requests made to the database.
func NewCommandMonitor(sensor *instana.Sensor) *event.CommandMonitor {
	mon := &commandMonitor{
		sensor: sensor,
		spans:  newSpanRegistry(),
	}

	return &event.CommandMonitor{
		Started:   mon.Started,
		Succeeded: mon.Succeeded,
		Failed:    mon.Failed,
	}
}

// Started traces command start initiating a new span. This span is finalized whenever either
// Succeeded() or Failed() method is called with an event containing the same RequestID.
func (m *commandMonitor) Started(ctx context.Context, evt *event.CommandStartedEvent) {
	parent, ok := instana.SpanFromContext(ctx)
	if !ok {
		return
	}

	ns := evt.DatabaseName
	if collection, ok := evt.Command.Lookup(evt.CommandName).StringValueOK(); ok {
		ns += "." + collection
	}

	sp := m.sensor.Tracer().StartSpan(
		"mongo",
		opentracing.ChildOf(parent.Context()),
		m.extractSpanTags(evt),
	)

	m.spans.Add(evt.RequestID, sp)
}

// Succeeded finalizes the command span started by Started()
func (m *commandMonitor) Succeeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	sp, ok := m.spans.Remove(evt.RequestID)
	if !ok {
		return
	}

	sp.Finish()
}

// Failed finalizes the command span started by Started() and logs the failure reason
func (m *commandMonitor) Failed(ctx context.Context, evt *event.CommandFailedEvent) {
	sp, ok := m.spans.Remove(evt.RequestID)
	if !ok {
		return
	}

	sp.SetTag("mongo.error", evt.Failure)
	sp.LogFields(otlog.Object("error", evt.Failure))

	sp.Finish()
}

func (m *commandMonitor) extractSpanTags(evt *event.CommandStartedEvent) opentracing.Tags {
	ns := evt.DatabaseName
	if collection, ok := evt.Command.Lookup(evt.CommandName).StringValueOK(); ok {
		ns += "." + collection
	}

	tags := opentracing.Tags{
		"mongo.service":   extractAddress(evt.ConnectionID),
		"mongo.namespace": ns,
		"mongo.command":   evt.CommandName,
	}

	if doc := evt.Command.Lookup("query"); doc.Validate() == nil {
		if data, err := bsonToJSON(doc); err != nil {
			m.sensor.Logger().Warn("failed to marshal mongodb ", evt.CommandName, " query to json: ", err)
		} else {
			tags["mongo.query"] = string(data)
		}
	}

	if doc := evt.Command.Lookup("filter"); doc.Validate() == nil {
		if data, err := bsonToJSON(evt.Command.Lookup("filter")); err != nil {
			m.sensor.Logger().Warn("failed to marshal mongodb ", evt.CommandName, " filter to json: ", err)
		} else {
			tags["mongo.filter"] = string(data)
		}
	}

	if doc, ok := extractCommandDocument(evt.CommandName, evt.Command); ok {
		if data, err := bsonToJSON(doc); err != nil {
			m.sensor.Logger().Warn("failed to marshal mongodb ", evt.CommandName, " document to json: ", err)
		} else {
			tags["mongo.json"] = string(data)
		}
	}

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

func extractCommandDocument(cmdName string, cmdBody bson.Raw) (bson.RawValue, bool) {
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

		if mapV := cmdBody.Lookup("map"); mapV.Validate() == nil {
			if s, ok := stringOrJavaScriptOK(mapV); ok {
				doc["map"] = s
			}
		}

		if reduceV := cmdBody.Lookup("reduce"); reduceV.Validate() == nil {
			if s, ok := stringOrJavaScriptOK(reduceV); ok {
				doc["reduce"] = s
			}
		}

		typ, data, err := bson.MarshalValue(doc)
		if err != nil {
			return bson.RawValue{}, false
		}

		v = bson.RawValue{
			Type:  typ,
			Value: data,
		}
	}

	return v, v.Validate() == nil
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
