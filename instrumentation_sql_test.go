// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstrumentSQLDriver(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, recorder))

	defer instana.ShutdownSensor()

	instana.InstrumentSQLDriver(s, "test_register_driver", sqlDriver{})
	assert.NotPanics(t, func() {
		instana.InstrumentSQLDriver(s, "test_register_driver", sqlDriver{})
	})
}

func TestOpenSQLDB(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, recorder))
	defer instana.ShutdownSensor()

	instana.InstrumentSQLDriver(s, "test_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "test_driver_with_instana")

	db, err := instana.SQLOpen("test_driver", "connection string")
	require.NoError(t, err)

	t.Run("Exec", func(t *testing.T) {
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
	})

	t.Run("Query", func(t *testing.T) {
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
	})
}

func TestOpenSQLDB_URIConnString(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, recorder))
	defer instana.ShutdownSensor()

	instana.InstrumentSQLDriver(s, "fake_db_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "test_driver_with_instana")

	db, err := instana.SQLOpen("fake_db_driver", "db://user1:p@55w0rd@db-host:1234/test-schema?param=value")
	require.NoError(t, err)

	_, err = db.Exec("TEST QUERY")
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	require.IsType(t, instana.SDKSpanData{}, spans[0].Data)
	data := spans[0].Data.(instana.SDKSpanData)

	assert.Equal(t, instana.SDKSpanTags{
		Name: "sdk.database",
		Type: "exit",
		Custom: map[string]interface{}{
			"tags": ot.Tags{
				"span.kind":     ext.SpanKindRPCClientEnum,
				"db.instance":   "test-schema",
				"db.statement":  "TEST QUERY",
				"db.type":       "sql",
				"peer.address":  "db://user1@db-host:1234/test-schema?param=value",
				"peer.hostname": "db-host",
				"peer.port":     "1234",
			},
		},
	}, data.Tags)
}

func TestOpenSQLDB_PostgresKVConnString(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, recorder))
	defer instana.ShutdownSensor()

	instana.InstrumentSQLDriver(s, "fake_postgres_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "fake_postgres_driver_with_instana")

	db, err := instana.SQLOpen("fake_postgres_driver", "host=db-host1,db-host-2 hostaddr=1.2.3.4,2.3.4.5 connect_timeout=10  port=1234 user=user1 password=p@55w0rd dbname=test-schema")
	require.NoError(t, err)

	_, err = db.Exec("TEST QUERY")
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	require.IsType(t, instana.SDKSpanData{}, spans[0].Data)
	data := spans[0].Data.(instana.SDKSpanData)

	assert.Equal(t, instana.SDKSpanTags{
		Name: "sdk.database",
		Type: "exit",
		Custom: map[string]interface{}{
			"tags": ot.Tags{
				"span.kind":     ext.SpanKindRPCClientEnum,
				"db.instance":   "test-schema",
				"db.statement":  "TEST QUERY",
				"db.type":       "sql",
				"peer.address":  "host=db-host1,db-host-2 hostaddr=1.2.3.4,2.3.4.5 connect_timeout=10  port=1234 user=user1 dbname=test-schema",
				"peer.hostname": "1.2.3.4,2.3.4.5",
				"peer.port":     "1234",
			},
		},
	}, data.Tags)
}

func TestOpenSQLDB_MySQLKVConnString(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, recorder))
	defer instana.ShutdownSensor()

	instana.InstrumentSQLDriver(s, "fake_mysql_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "fake_mysql_driver_with_instana")

	db, err := instana.SQLOpen("fake_mysql_driver", "Server=db-host1, db-host2;Database=test-schema;Port=1234;Uid=user1;Pwd=p@55w0rd;")
	require.NoError(t, err)

	_, err = db.Exec("TEST QUERY")
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	require.IsType(t, instana.SDKSpanData{}, spans[0].Data)
	data := spans[0].Data.(instana.SDKSpanData)

	assert.Equal(t, instana.SDKSpanTags{
		Name: "sdk.database",
		Type: "exit",
		Custom: map[string]interface{}{
			"tags": ot.Tags{
				"span.kind":     ext.SpanKindRPCClientEnum,
				"db.instance":   "test-schema",
				"db.statement":  "TEST QUERY",
				"db.type":       "sql",
				"peer.address":  "Server=db-host1, db-host2;Database=test-schema;Port=1234;Uid=user1;",
				"peer.hostname": "db-host1, db-host2",
				"peer.port":     "1234",
			},
		},
	}, data.Tags)
}

func TestNoPanicWithNotParsableConnectionString(t *testing.T) {
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, instana.NewTestRecorder()))
	defer instana.ShutdownSensor()

	instana.InstrumentSQLDriver(s, "test_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "test_driver_with_instana")

	assert.NotPanics(t, func() {
		_, _ = instana.SQLOpen("test_driver",
			"postgres:mysecretpassword@localhost/postgres")
	})
}

func TestProcedureWithCheckerOnStmt(t *testing.T) {
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, instana.NewTestRecorder()))
	defer instana.ShutdownSensor()

	ch := make(chan int, 2)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		// Here we make sure that stmt.CheckNamedValue from sqlDriver2 is called
		<-ch
		wg.Done()
	}()

	driver := sqlDriver2{
		ch: ch,
	}

	instana.InstrumentSQLDriver(s, "test_driver2", driver)
	db, err := instana.SQLOpen("test_driver2", "some datasource")

	assert.NoError(t, err)

	var outValue string
	_, err = db.Exec("CALL SOME_PROCEDURE(?)", sql.Out{Dest: &outValue})

	// Here we expect the instrumentation to look for the driver's conn.CheckNamedValue implementation.
	// If there is none, we check stmt.CheckNamedValue, which sqlDriver2 has.
	// If there is none, we return nil from our side, since driver.ErrSkip won't work for CheckNamedValue, as seen here:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.19.1:src/database/sql/driver/driver.go;l=143
	// and here: https://cs.opensource.google/go/go/+/refs/tags/go1.19.1:src/database/sql/driver/driver.go;l=399
	assert.NoError(t, err)

	wg.Wait()
}

func TestProcedureWithNoDefaultChecker(t *testing.T) {
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, instana.NewTestRecorder()))
	defer instana.ShutdownSensor()

	driver := sqlDriver3{}

	instana.InstrumentSQLDriver(s, "test_driver3", driver)
	db, err := instana.SQLOpen("test_driver3", "some datasource")

	assert.NoError(t, err)

	_, err = db.Exec("select $1", int32(1))

	// Here we expect the instrumentation to look for the driver's conn.CheckNamedValue implementation.
	// If there is none, we check stmt.CheckNamedValue, which sqlDriver also doesn't have.
	// If there is none, we return nil from our side, since driver.ErrSkip won't work for CheckNamedValue, as seen here:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.19.1:src/database/sql/driver/driver.go;l=143
	// and here: https://cs.opensource.google/go/go/+/refs/tags/go1.19.1:src/database/sql/driver/driver.go;l=399
	assert.NoError(t, err)
}

type sqlDriver struct{ Error error }

func (drv sqlDriver) Open(name string) (driver.Conn, error) { return sqlConn{drv.Error}, nil } //nolint:gosimple

type sqlConn struct{ Error error }

func (conn sqlConn) Prepare(query string) (driver.Stmt, error) { return sqlStmt{conn.Error}, nil } //nolint:gosimple
func (s sqlConn) Close() error                                 { return driver.ErrSkip }
func (s sqlConn) Begin() (driver.Tx, error)                    { return nil, driver.ErrSkip }

func (conn sqlConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return sqlResult{}, conn.Error
}

func (conn sqlConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return sqlRows{}, conn.Error
}

type sqlStmt struct{ Error error }

func (sqlStmt) Close() error                                         { return nil }
func (sqlStmt) NumInput() int                                        { return -1 }
func (stmt sqlStmt) Exec(args []driver.Value) (driver.Result, error) { return sqlResult{}, stmt.Error }
func (stmt sqlStmt) Query(args []driver.Value) (driver.Rows, error)  { return sqlRows{}, stmt.Error }

type sqlResult struct{}

func (sqlResult) LastInsertId() (int64, error) { return 42, nil }
func (sqlResult) RowsAffected() (int64, error) { return 100, nil }

type sqlRows struct{}

func (sqlRows) Columns() []string              { return []string{"col1", "col2"} }
func (sqlRows) Close() error                   { return nil }
func (sqlRows) Next(dest []driver.Value) error { return io.EOF }

// Driver use case:
// * driver.Conn doesn't implement Exec or ExecContext
// * driver.Conn doesn't implement the driver.NamedValueChecker interface (CheckNamedValue method)
// * driver.Stmt does implement the driver.NamedValueChecker interface (CheckNamedValue method)
// * Our wrapper ALWAYS implements ExecContext, no matter what

type sqlDriver2 struct {
	Error error
	ch    chan int
}

func (drv sqlDriver2) Open(name string) (driver.Conn, error) { return sqlConn2{drv.Error, drv.ch}, nil } //nolint:gosimple

type sqlConn2 struct {
	Error error
	ch    chan int
}

func (conn sqlConn2) Prepare(query string) (driver.Stmt, error) {
	return sqlStmt2{conn.Error, conn.ch}, nil //nolint:gosimple
}
func (s sqlConn2) Close() error { return driver.ErrSkip }

func (s sqlConn2) Begin() (driver.Tx, error) { return nil, driver.ErrSkip }

type sqlStmt2 struct {
	Error error
	ch    chan int
}

func (sqlStmt2) Close() error  { return nil }
func (sqlStmt2) NumInput() int { return -1 }
func (stmt sqlStmt2) Exec(args []driver.Value) (driver.Result, error) {
	return sqlResult2{}, stmt.Error
}

func (stmt sqlStmt2) Query(args []driver.Value) (driver.Rows, error) {
	return sqlRows2{}, stmt.Error
}

func (stmt sqlStmt2) CheckNamedValue(d *driver.NamedValue) error {
	stmt.ch <- 1
	return nil
}

type sqlResult2 struct{}

func (sqlResult2) LastInsertId() (int64, error) { return 42, nil }
func (sqlResult2) RowsAffected() (int64, error) { return 100, nil }

type sqlRows2 struct{}

func (sqlRows2) Columns() []string              { return []string{"col1", "col2"} }
func (sqlRows2) Close() error                   { return nil }
func (sqlRows2) Next(dest []driver.Value) error { return io.EOF }

// Driver use case: driver does not implement NamedValueChecker,arg type checking is internal.
// The idea is to mock pq: https://github.com/lib/pq/blob/8446d16b8935fdf2b5c0fe333538ac395e3e1e4b/encode.go#L31

type sqlDriver3 struct{ Error error }

func (drv sqlDriver3) Open(name string) (driver.Conn, error) { return sqlConn3{drv.Error}, nil } //nolint:gosimple

type sqlConn3 struct{ Error error }

func (conn sqlConn3) Prepare(query string) (driver.Stmt, error) { return sqlStmt3{conn.Error}, nil } //nolint:gosimple
func (s sqlConn3) Close() error                                 { return driver.ErrSkip }
func (s sqlConn3) Begin() (driver.Tx, error)                    { return nil, driver.ErrSkip }

func (conn sqlConn3) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	var err error

	if _, ok := args[0].Value.(int32); ok {
		err = errors.New("invalid type int32")
	}

	return sqlResult3{}, err
}

func (conn sqlConn3) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return sqlRows3{}, conn.Error
}

type sqlStmt3 struct{ Error error }

func (sqlStmt3) Close() error  { return nil }
func (sqlStmt3) NumInput() int { return -1 }
func (stmt sqlStmt3) Exec(args []driver.Value) (driver.Result, error) {
	return sqlResult3{}, stmt.Error
}
func (stmt sqlStmt3) Query(args []driver.Value) (driver.Rows, error) { return sqlRows3{}, stmt.Error }

type sqlResult3 struct{}

func (sqlResult3) LastInsertId() (int64, error) { return 42, nil }
func (sqlResult3) RowsAffected() (int64, error) { return 100, nil }

type sqlRows3 struct{}

func (sqlRows3) Columns() []string              { return []string{"col1", "col2"} }
func (sqlRows3) Close() error                   { return nil }
func (sqlRows3) Next(dest []driver.Value) error { return io.EOF }
