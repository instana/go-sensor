// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

// +build go1.10

package instana_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func TestWrapSQLConnector_Exec(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service: "go-sensor-test",
	}, recorder))

	db := sql.OpenDB(instana.WrapSQLConnector(s, "connection string", sqlConnector{}))

	res, err := db.Exec("TEST QUERY")
	require.NoError(t, err)

	lastID, err := res.LastInsertId()
	require.NoError(t, err)
	assert.Equal(t, int64(42), lastID)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, instana.SDKSpanTags{
		Name: "sdk.database",
		Type: "exit",
		Custom: map[string]interface{}{
			"tags": ot.Tags{
				"span.kind":    ext.SpanKindRPCClientEnum,
				"db.instance":  "connection string",
				"db.statement": "TEST QUERY",
				"db.type":      "sql",
				"peer.address": "connection string",
			},
		},
	}, data.Tags)
}

func TestWrapSQLConnector_Exec_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service: "go-sensor-test",
	}, recorder))

	dbErr := errors.New("something went wrong")
	db := sql.OpenDB(instana.WrapSQLConnector(s, "connection string", sqlConnector{
		Error: dbErr,
	}))

	_, err := db.Exec("TEST QUERY")
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	require.IsType(t, map[uint64]map[string]interface{}{}, data.Tags.Custom["logs"])
	logs := data.Tags.Custom["logs"].(map[uint64]map[string]interface{})

	collected := make(map[string][]interface{})
	for _, l := range logs {
		for k, v := range l {
			if k == "error.object" {
				k = "error"
			}
			collected[k] = append(collected[k], v)
		}
	}
	assert.Contains(t, collected["error"], dbErr)
}

func TestWrapSQLConnector_Query(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service: "go-sensor-test",
	}, recorder))

	db := sql.OpenDB(instana.WrapSQLConnector(s, "connection string", sqlConnector{}))

	res, err := db.Query("TEST QUERY")
	require.NoError(t, err)

	cols, err := res.Columns()
	require.NoError(t, err)
	assert.Equal(t, []string{"col1", "col2"}, cols)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, instana.SDKSpanTags{
		Name: "sdk.database",
		Type: "exit",
		Custom: map[string]interface{}{
			"tags": ot.Tags{
				"span.kind":    ext.SpanKindRPCClientEnum,
				"db.instance":  "connection string",
				"db.statement": "TEST QUERY",
				"db.type":      "sql",
				"peer.address": "connection string",
			},
		},
	}, data.Tags)
}

func TestWrapSQLConnector_Query_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service: "go-sensor-test",
	}, recorder))

	dbErr := errors.New("something went wrong")
	db := sql.OpenDB(instana.WrapSQLConnector(s, "connection string", sqlConnector{
		Error: dbErr,
	}))

	_, err := db.Query("TEST QUERY")
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	require.IsType(t, map[uint64]map[string]interface{}{}, data.Tags.Custom["logs"])
	logs := data.Tags.Custom["logs"].(map[uint64]map[string]interface{})

	collected := make(map[string][]interface{})
	for _, l := range logs {
		for k, v := range l {
			if k == "error.object" {
				k = "error"
			}
			collected[k] = append(collected[k], v)
		}
	}
	assert.Contains(t, collected["error"], dbErr)
}

type sqlConnector struct{ Error error }

func (c sqlConnector) Connect(context.Context) (driver.Conn, error) { return sqlConn{c.Error}, nil } //nolint:gosimple
func (sqlConnector) Driver() driver.Driver                          { return sqlDriver{} }
