// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"os"
	"testing"
	_ "unsafe"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstrumentSQLDriver(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	instana.InstrumentSQLDriver(c, "test_register_driver", sqlDriver{})
	assert.NotPanics(t, func() {
		instana.InstrumentSQLDriver(c, "test_register_driver", sqlDriver{})
	})
}

func BenchmarkSQLOpenAndExec(b *testing.B) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	instana.InstrumentSQLDriver(c, "test_driver", sqlDriver{})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db, err := instana.SQLOpen("test_driver", "connection string")
		if err != nil {
			b.Fatal(err)
		}
		_, err = db.Exec("TEST QUERY")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestOpenSQLDB_WithoutParentSpan(t *testing.T) {

	os.Setenv("INSTANA_ALLOW_ROOT_EXIT_SPAN", "1")
	defer os.Unsetenv("INSTANA_ALLOW_ROOT_EXIT_SPAN")

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	instana.InstrumentSQLDriver(c, "test_driver_without_parent_span", sqlDriver{})
	require.Contains(t, sql.Drivers(), "test_driver_without_parent_span_with_instana")

	db, err := instana.SQLOpen("test_driver_without_parent_span", "connection string")
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

func TestOpenSQLDB(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}
	instana.InstrumentSQLDriver(c, "test_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "test_driver_with_instana")

	db, err := instana.SQLOpen("test_driver", "connection string")
	require.NoError(t, err)

	t.Run("Exec", func(t *testing.T) {
		res, err := db.ExecContext(ctx, "TEST QUERY")
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
		res, err := db.QueryContext(ctx, "TEST QUERY")
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

func TestPostgresDB(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}
	instana.InstrumentSQLDriver(c, "pg_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "pg_driver_with_instana")

	db, err := instana.SQLOpen("pg_driver",
		"host=db-host1,db-host-2 hostaddr=1.2.3.4,2.3.4.5 connect_timeout=10  port=1234 user=user1 password=p@55w0rd dbname=test-schema")
	require.NoError(t, err)

	tests := []struct {
		name string
		op   func()
		want instana.PostgreSQLSpanTags
	}{
		{
			name: "exec",
			op: func() {
				res, err := db.ExecContext(ctx, "TEST QUERY")
				require.NoError(t, err)

				lastID, err := res.LastInsertId()
				require.NoError(t, err)
				assert.Equal(t, int64(42), lastID)
			},
			want: instana.PostgreSQLSpanTags{
				Host:  "1.2.3.4,2.3.4.5",
				DB:    "test-schema",
				Port:  "1234",
				User:  "user1",
				Stmt:  "TEST QUERY",
				Error: "",
			},
		},
		{
			name: "query context",
			op: func() {
				res, err := db.QueryContext(ctx, "TEST QUERY")
				require.NoError(t, err)

				cols, err := res.Columns()
				require.NoError(t, err)
				assert.Equal(t, []string{"col1", "col2"}, cols)
			},
			want: instana.PostgreSQLSpanTags{
				Host:  "1.2.3.4,2.3.4.5",
				DB:    "test-schema",
				Port:  "1234",
				User:  "user1",
				Stmt:  "TEST QUERY",
				Error: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.op()

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			span := spans[0]
			assert.Equal(t, 0, span.Ec)
			assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

			require.IsType(t, instana.PostgreSQLSpanData{}, span.Data)
			data := span.Data.(instana.PostgreSQLSpanData)

			assert.Equal(t, tt.want, data.Tags)
		})
	}
}

func TestCouchbaseDB(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	tests := []struct {
		name string
		conn instana.DbConnDetails
		want instana.CouchbaseSpanTags
	}{
		{
			name: "exec",
			conn: instana.DbConnDetails{
				DatabaseName: instana.Couchbase,
				RawString:    "127.0.0.1",
			},
			want: instana.CouchbaseSpanTags{
				Bucket: "",
				Host:   "127.0.0.1",
				Type:   "",
				SQL:    "TEST QUERY",
				Error:  "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sp, dbKey := instana.StartSQLSpan(ctx, tt.conn, "TEST QUERY", c)
			sp.Finish()

			assert.Equal(t, dbKey, "couchbase")

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			span := spans[0]
			assert.Equal(t, 0, span.Ec)
			assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

			require.IsType(t, instana.CouchbaseSpanData{}, span.Data)
			data := span.Data.(instana.CouchbaseSpanData)

			assert.Equal(t, tt.want, data.Tags)
		})
	}
}

func TestCosmos(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	tests := []struct {
		name string
		tags map[string]string
		conn instana.DbConnDetails
		want instana.CosmosSpanTags
	}{
		{
			name: "exec",
			conn: instana.DbConnDetails{
				DatabaseName: instana.Cosmos,
			},
			tags: map[string]string{
				"cosmos.con": "https://test.com:443/",
				"cosmos.db":  "test-db",
			},
			want: instana.CosmosSpanTags{
				ConnectionURL: "https://test.com:443/",
				Database:      "test-db",
				Type:          "",
				Sql:           "TEST QUERY",
				Object:        "",
				PartitionKey:  "",
				ReturnCode:    "",
				Error:         "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sp, dbKey := instana.StartSQLSpan(ctx, tt.conn, "TEST QUERY", c)
			for key, val := range tt.tags {
				sp.SetTag(key, val)
			}
			sp.Finish()

			assert.Equal(t, dbKey, "cosmos")

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			span := spans[0]
			assert.Equal(t, 0, span.Ec)
			assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

			require.IsType(t, instana.CosmosSpanData{}, span.Data)
			data := span.Data.(instana.CosmosSpanData)

			assert.Equal(t, tt.want, data.Tags)
		})
	}
}

func TestOpenDB2(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}
	instana.InstrumentSQLDriver(c, "go_ibm_db", sqlDriver{})
	require.Contains(t, sql.Drivers(), "go_ibm_db_with_instana")

	testcases := map[string]struct {
		DSN            string
		DriverName     string
		ExpectedConfig instana.SDKSpanTags
	}{
		"DB2_With_Server_Field": {
			DSN: "Server=localhost:50000;DATABASE=sample;UID=db2inst1;PWD=password",
			ExpectedConfig: instana.SDKSpanTags{
				Name: "sdk.database",
				Type: "exit",
				Custom: map[string]interface{}{
					"tags": ot.Tags{
						"span.kind":     ext.SpanKindRPCClientEnum,
						"db.instance":   "sample",
						"db.statement":  "TEST QUERY",
						"db.type":       "sql",
						"peer.address":  "Server=localhost:50000;DATABASE=sample;UID=db2inst1;",
						"peer.hostname": "localhost",
						"peer.port":     "50000",
					},
				},
			},
		},
		"DB2_With_No_Port_In_Server_Field": {
			DSN: "Server=localhost;DATABASE=sample;UID=db2inst1;PWD=password",
			ExpectedConfig: instana.SDKSpanTags{
				Name: "sdk.database",
				Type: "exit",
				Custom: map[string]interface{}{
					"tags": ot.Tags{
						"span.kind":     ext.SpanKindRPCClientEnum,
						"db.instance":   "sample",
						"db.statement":  "TEST QUERY",
						"db.type":       "sql",
						"peer.address":  "Server=localhost;DATABASE=sample;UID=db2inst1;",
						"peer.hostname": "localhost",
					},
				},
			},
		},
		"DB2_With_Hostname_And_Port": {
			DSN: "Hostname=localhost;Port=50000;DATABASE=sample;UID=db2inst1;PWD=password",
			ExpectedConfig: instana.SDKSpanTags{
				Name: "sdk.database",
				Type: "exit",
				Custom: map[string]interface{}{
					"tags": ot.Tags{
						"span.kind":     ext.SpanKindRPCClientEnum,
						"db.instance":   "sample",
						"db.statement":  "TEST QUERY",
						"db.type":       "sql",
						"peer.address":  "Hostname=localhost;Port=50000;DATABASE=sample;UID=db2inst1;",
						"peer.hostname": "localhost",
						"peer.port":     "50000",
					},
				},
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			db, err := instana.SQLOpen("go_ibm_db", testcase.DSN)
			require.NoError(t, err)

			res, err := db.ExecContext(ctx, "TEST QUERY")
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

			assert.Equal(t, testcase.ExpectedConfig, data.Tags)
		})
	}

	// without known driver name, tracer will find IBM Db2 as a mysql db.
	// It is STRONGLY recommended that the customer should provide driver name
	// as `go_ibm_db` for Db2 database.
	// The mismatch in host and port is expected if driver name is not provided.
	t.Run("DB2_Without_unknown_driver_name", func(t *testing.T) {

		instana.InstrumentSQLDriver(c, "unknown_db2_driver", sqlDriver{})
		require.Contains(t, sql.Drivers(), "unknown_db2_driver_with_instana")

		db, err := instana.SQLOpen("unknown_db2_driver", "Server=localhost:50000;DATABASE=sample;UID=db2inst1;PWD=password")
		require.NoError(t, err)

		res, err := db.ExecContext(ctx, "TEST QUERY")
		require.NoError(t, err)

		lastID, err := res.LastInsertId()
		require.NoError(t, err)
		assert.Equal(t, int64(42), lastID)

		spans := recorder.GetQueuedSpans()
		require.Len(t, spans, 1)

		span := spans[0]
		assert.Equal(t, 0, span.Ec)
		assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

		require.IsType(t, instana.MySQLSpanData{}, span.Data)
		data := span.Data.(instana.MySQLSpanData)

		assert.Equal(t, instana.MySQLSpanTags{
			Host:  "localhost:50000",
			Port:  "",
			DB:    "sample",
			User:  "db2inst1",
			Stmt:  "TEST QUERY",
			Error: "",
		}, data.Tags)
	})
}

func TestDSNParing(t *testing.T) {
	testcases := map[string]struct {
		DSN            string
		DriverName     string
		ExpectedConfig instana.DbConnDetails
	}{
		"URI": {
			DSN: "db://user1:p@55w0rd@db-host:1234/test-schema?param=value",
			ExpectedConfig: instana.DbConnDetails{
				Schema:    "test-schema",
				RawString: "db://user1@db-host:1234/test-schema?param=value",
				Host:      "db-host",
				Port:      "1234",
				User:      "user1",
			},
		},
		"Postgres": {
			DSN: "host=db-host1,db-host-2 hostaddr=1.2.3.4,2.3.4.5 connect_timeout=10  port=1234 user=user1 password=p@55w0rd dbname=test-schema",
			ExpectedConfig: instana.DbConnDetails{
				RawString:    "host=db-host1,db-host-2 hostaddr=1.2.3.4,2.3.4.5 connect_timeout=10  port=1234 user=user1 dbname=test-schema",
				Host:         "1.2.3.4,2.3.4.5",
				Port:         "1234",
				Schema:       "test-schema",
				User:         "user1",
				DatabaseName: "postgres",
			},
		},
		"Postgres - URI - with password": {
			DSN: "postgresql://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable",
			ExpectedConfig: instana.DbConnDetails{
				Schema:       "postgres",
				RawString:    "postgresql://postgres@localhost:5432/postgres?sslmode=disable",
				Host:         "localhost",
				Port:         "5432",
				User:         "postgres",
				DatabaseName: "postgres",
			},
		},
		"Postgres - URI - without password": {
			DSN: "postgresql://postgres@localhost:5432/postgres?sslmode=disable",
			ExpectedConfig: instana.DbConnDetails{
				Schema:       "postgres",
				RawString:    "postgresql://postgres@localhost:5432/postgres?sslmode=disable",
				Host:         "localhost",
				Port:         "5432",
				User:         "postgres",
				DatabaseName: "postgres",
			},
		},
		"MySQL": {
			DSN: "Server=db-host1, db-host2;Database=test-schema;Port=1234;Uid=user1;Pwd=p@55w0rd;",
			ExpectedConfig: instana.DbConnDetails{
				RawString:    "Server=db-host1, db-host2;Database=test-schema;Port=1234;Uid=user1;",
				Host:         "db-host1, db-host2",
				Port:         "1234",
				Schema:       "test-schema",
				User:         "user1",
				DatabaseName: "mysql",
			},
		},
		"MySQL DSN": {
			DSN: "go:gopw@tcp(localhost:3306)/godb",
			ExpectedConfig: instana.DbConnDetails{
				RawString:    "go:gopw@tcp(localhost:3306)/godb",
				Host:         "localhost",
				Port:         "3306",
				Schema:       "godb",
				User:         "go",
				DatabaseName: "mysql",
			},
		},
		"MySQL DSN - without port": {
			DSN: "go:gopw@tcp(localhost)/godb",
			ExpectedConfig: instana.DbConnDetails{
				RawString:    "go:gopw@tcp(localhost)/godb",
				Host:         "localhost",
				Port:         "3306",
				Schema:       "godb",
				User:         "go",
				DatabaseName: "mysql",
			},
		},
		"MySQL DSN - with params": {
			DSN: "go:gopw@tcp(localhost:3306)/godb?parseTime=true",
			ExpectedConfig: instana.DbConnDetails{
				RawString:    "go:gopw@tcp(localhost:3306)/godb?parseTime=true",
				Host:         "localhost",
				Port:         "3306",
				Schema:       "godb",
				User:         "go",
				DatabaseName: "mysql",
			},
		},
		"Redis_full_conn_string": {
			DSN: "user:p455w0rd@127.0.0.15:3679",
			ExpectedConfig: instana.DbConnDetails{
				RawString:    "user:p455w0rd@127.0.0.15:3679",
				Host:         "127.0.0.15",
				Port:         "3679",
				Schema:       "",
				User:         "",
				DatabaseName: "redis",
			},
		},
		"Redis_no_user": {
			DSN: ":p455w0rd@127.0.0.15:3679",
			ExpectedConfig: instana.DbConnDetails{
				RawString:    ":p455w0rd@127.0.0.15:3679",
				Host:         "127.0.0.15",
				Port:         "3679",
				Schema:       "",
				User:         "",
				DatabaseName: "redis",
			},
		},
		"SQLite": {
			DSN: "/home/user/products.db",
			ExpectedConfig: instana.DbConnDetails{
				RawString: "/home/user/products.db",
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			connDetails := instana.ParseDBConnDetails(testcase.DSN)
			assert.Equal(t, testcase.ExpectedConfig, connDetails)
		})
	}
}

func TestOpenSQLDB_URIConnString(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	instana.InstrumentSQLDriver(c, "fake_db_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "test_driver_with_instana")

	db, err := instana.SQLOpen("fake_db_driver", "db://user1:p@55w0rd@db-host:1234/test-schema?param=value")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "TEST QUERY")
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
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	instana.InstrumentSQLDriver(c, "fake_postgres_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "fake_postgres_driver_with_instana")

	db, err := instana.SQLOpen("fake_postgres_driver", "host=db-host1,db-host-2 hostaddr=1.2.3.4,2.3.4.5 connect_timeout=10  port=1234 user=user1 password=p@55w0rd dbname=test-schema")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "TEST QUERY")
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	require.IsType(t, instana.PostgreSQLSpanData{}, spans[0].Data)
	data := spans[0].Data.(instana.PostgreSQLSpanData)

	assert.Equal(t, instana.PostgreSQLSpanTags{
		Host:  "1.2.3.4,2.3.4.5",
		DB:    "test-schema",
		Port:  "1234",
		User:  "user1",
		Stmt:  "TEST QUERY",
		Error: "",
	}, data.Tags)
}

func TestOpenSQLDB_MySQLKVConnString(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	instana.InstrumentSQLDriver(c, "fake_mysql_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "fake_mysql_driver_with_instana")

	db, err := instana.SQLOpen("fake_mysql_driver", "Server=db-host1, db-host2;Database=test-schema;Port=1234;Uid=user1;Pwd=p@55w0rd;")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "TEST QUERY")
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	require.IsType(t, instana.MySQLSpanData{}, spans[0].Data)
	data := spans[0].Data.(instana.MySQLSpanData)

	assert.Equal(t, instana.MySQLSpanTags{
		Host:  "db-host1, db-host2",
		Port:  "1234",
		DB:    "test-schema",
		User:  "user1",
		Stmt:  "TEST QUERY",
		Error: "",
	}, data.Tags)
}

func TestOpenSQLDB_RedisConnString(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	instana.InstrumentSQLDriver(c, "fake_redis_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "fake_redis_driver_with_instana")

	db, err := instana.SQLOpen("fake_redis_driver", ":p455w0rd@192.168.2.10:6790")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "SET name Instana EX 15")
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	require.IsType(t, instana.RedisSpanData{}, spans[0].Data)
	data := spans[0].Data.(instana.RedisSpanData)

	assert.Equal(t, instana.RedisSpanTags{
		Connection: "192.168.2.10:6790",
		Command:    "SET",
		Error:      "",
	}, data.Tags)
}

func TestConnPrepareContext(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	instana.InstrumentSQLDriver(c, "fake_pc", sqlDriver{})
	require.Contains(t, sql.Drivers(), "fake_pc_with_instana")

	db, err := instana.SQLOpen("fake_pc", "conn string")
	require.NoError(t, err)

	stmt, err := db.PrepareContext(ctx, "select 1 from table")
	require.NoError(t, err)

	_, err = stmt.QueryContext(ctx)
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
				"span.kind":    ext.SpanKindRPCClientEnum,
				"db.instance":  "conn string",
				"db.statement": "select 1 from table",
				"db.type":      "sql",
				"peer.address": "conn string",
			},
		},
	}, data.Tags)
}

func TestConnPrepareContextWithError(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	instana.InstrumentSQLDriver(c, "fake_conn_pc_error", sqlDriver{Error: errors.New("some error")})
	require.Contains(t, sql.Drivers(), "fake_conn_pc_error_with_instana")

	db, err := instana.SQLOpen("fake_conn_pc_error", "conn string")
	require.NoError(t, err)

	stmt, err := db.PrepareContext(ctx, "select 1 from table")
	require.NoError(t, err)

	_, err = stmt.QueryContext(ctx)
	require.Error(t, err)

	spans := recorder.GetQueuedSpans()

	require.Len(t, spans, 2)

	assert.Equal(t, spans[0].Ec, 1)

	require.IsType(t, instana.SDKSpanData{}, spans[0].Data)
	data := spans[0].Data.(instana.SDKSpanData)

	assert.Equal(t, instana.SDKSpanTags{
		Name: "sdk.database",
		Type: "exit",
		Custom: map[string]interface{}{
			"tags": ot.Tags{
				"span.kind":    ext.SpanKindRPCClientEnum,
				"db.error":     "some error",
				"db.instance":  "conn string",
				"db.statement": "select 1 from table",
				"db.type":      "sql",
				"peer.address": "conn string",
			},
		},
	}, data.Tags)
}

func TestStmtExecContext(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	instana.InstrumentSQLDriver(c, "fake_stmt_ec", sqlDriver{})
	require.Contains(t, sql.Drivers(), "fake_stmt_ec_with_instana")

	db, err := instana.SQLOpen("fake_stmt_ec", "conn string")
	require.NoError(t, err)

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	stmt, err := db.PrepareContext(ctx, "select 1 from table")
	require.NoError(t, err)

	_, err = stmt.ExecContext(ctx)
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
				"span.kind":    ext.SpanKindRPCClientEnum,
				"db.instance":  "conn string",
				"db.statement": "select 1 from table",
				"db.type":      "sql",
				"peer.address": "conn string",
			},
		},
	}, data.Tags)
}

func TestStmtExecContextWithError(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	instana.InstrumentSQLDriver(c, "fake_stmt_ec_with_error", sqlDriver{Error: errors.New("oh no")})
	require.Contains(t, sql.Drivers(), "fake_stmt_ec_with_error_with_instana")

	db, err := instana.SQLOpen("fake_stmt_ec_with_error", "conn string")
	require.NoError(t, err)

	stmt, err := db.PrepareContext(ctx, "select 1 from table")
	require.NoError(t, err)

	_, err = stmt.ExecContext(ctx)
	require.Error(t, err)

	spans := recorder.GetQueuedSpans()

	require.Len(t, spans, 2)

	require.IsType(t, instana.SDKSpanData{}, spans[0].Data)
	data := spans[0].Data.(instana.SDKSpanData)

	assert.Equal(t, instana.SDKSpanTags{
		Name: "sdk.database",
		Type: "exit",
		Custom: map[string]interface{}{
			"tags": ot.Tags{
				"span.kind":    ext.SpanKindRPCClientEnum,
				"db.error":     "oh no",
				"db.instance":  "conn string",
				"db.statement": "select 1 from table",
				"db.type":      "sql",
				"peer.address": "conn string",
			},
		},
	}, data.Tags)
}

func TestConnPrepareContextWithErrorOnReturn(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	instana.InstrumentSQLDriver(c, "fake_conn_pc_error_on_ret", sqlDriver{PrepareError: errors.New("oh no")})
	require.Contains(t, sql.Drivers(), "fake_conn_pc_error_on_ret_with_instana")

	db, err := instana.SQLOpen("fake_conn_pc_error_on_ret", "conn string")
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.PrepareContext(ctx, "select 1 from table")
	require.Error(t, err)
}

func TestOpenSQLDB_RedisConnString_WithError(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	instana.InstrumentSQLDriver(c, "fake_redis_driver_with_error", sqlDriver{Error: errors.New("oops")})
	require.Contains(t, sql.Drivers(), "fake_redis_driver_with_error_with_instana")

	db, err := instana.SQLOpen("fake_redis_driver_with_error", ":p455w0rd@192.168.2.10:6790")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "SET name Instana EX 15")

	require.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	require.IsType(t, instana.RedisSpanData{}, spans[0].Data)
	data := spans[0].Data.(instana.RedisSpanData)

	assert.Equal(t, instana.RedisSpanTags{
		Connection: "192.168.2.10:6790",
		Command:    "SET",
		Error:      "oops",
	}, data.Tags)
}

func TestOpenSQLDB_RedisKVConnString(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	instana.InstrumentSQLDriver(c, "fake_redis_kv_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "fake_redis_kv_driver_with_instana")

	db, err := instana.SQLOpen("fake_redis_kv_driver", "192.168.2.10:6790")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "SET name Instana EX 15")
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	require.IsType(t, instana.RedisSpanData{}, spans[0].Data)
	data := spans[0].Data.(instana.RedisSpanData)

	assert.Equal(t, instana.RedisSpanTags{
		Connection: "192.168.2.10:6790",
		Command:    "SET",
		Error:      "",
	}, data.Tags)
}

func TestStmtExecContext_WithRedisCommands(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	instana.InstrumentSQLDriver(c, "fake_redis_driver_2", sqlDriver{})
	require.Contains(t, sql.Drivers(), "fake_redis_driver_2_with_instana")

	t.Run("valid redis command", func(t *testing.T) {

		db, err := instana.SQLOpen("fake_redis_driver_2", "192.168.2.10:6790")
		require.NoError(t, err)

		defer db.Close()

		_, err = db.ExecContext(ctx, "GET key")
		require.NoError(t, err)

		spans := recorder.GetQueuedSpans()
		require.Len(t, spans, 1)

		require.IsType(t, instana.RedisSpanData{}, spans[0].Data)
		data := spans[0].Data.(instana.RedisSpanData)

		assert.Equal(t, instana.RedisSpanTags{
			Connection: "192.168.2.10:6790",
			Command:    "GET",
			Error:      "",
		}, data.Tags)
	})

	t.Run("With multi word command", func(t *testing.T) {

		db, err := instana.SQLOpen("fake_redis_driver_2", "192.168.2.10:6790")
		require.NoError(t, err)

		defer db.Close()

		_, err = db.ExecContext(ctx, "CONFIG GET key")
		require.NoError(t, err)

		spans := recorder.GetQueuedSpans()
		require.Len(t, spans, 1)

		require.IsType(t, instana.RedisSpanData{}, spans[0].Data)
		data := spans[0].Data.(instana.RedisSpanData)

		assert.Equal(t, instana.RedisSpanTags{
			Connection: "192.168.2.10:6790",
			Command:    "CONFIG GET",
			Error:      "",
		}, data.Tags)
	})

	t.Run("wrong/unknown(to go sensor) redis command", func(t *testing.T) {

		db, err := instana.SQLOpen("fake_redis_driver_2", "192.168.2.10:6790")
		require.NoError(t, err)

		defer db.Close()

		_, err = db.ExecContext(ctx, "SELECT key")
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
					"span.kind":    ext.SpanKindRPCClientEnum,
					"db.instance":  "192.168.2.10:6790",
					"db.statement": "SELECT key",
					"db.type":      "sql",
					"peer.address": "192.168.2.10:6790",
				},
			},
		}, data.Tags)
	})

	t.Run("empty query", func(t *testing.T) {

		db, err := instana.SQLOpen("fake_redis_driver_2", "192.168.2.10:6790")
		require.NoError(t, err)

		defer db.Close()

		_, err = db.ExecContext(ctx, "")
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
					"span.kind":    ext.SpanKindRPCClientEnum,
					"db.instance":  "192.168.2.10:6790",
					"db.statement": "",
					"db.type":      "sql",
					"peer.address": "192.168.2.10:6790",
				},
			},
		}, data.Tags)
	})

	t.Run("transaction", func(t *testing.T) {

		db, err := instana.SQLOpen("fake_redis_driver_2", "192.168.2.10:6790")
		require.NoError(t, err)

		defer db.Close()

		_, err = db.ExecContext(ctx, "MULTI")
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, "SET", "key1", "value1")
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, "INCR", "counter")
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, "EXEC")
		require.NoError(t, err)

		spans := recorder.GetQueuedSpans()
		require.Len(t, spans, 4)

		testcases := []struct {
			Command string
		}{
			{
				Command: "MULTI",
			},
			{
				Command: "SET",
			},
			{
				Command: "INCR",
			},
			{
				Command: "EXEC",
			},
		}

		for i, tc := range testcases {
			require.IsType(t, instana.RedisSpanData{}, spans[i].Data)
			data := spans[i].Data.(instana.RedisSpanData)

			assert.Equal(t, instana.RedisSpanTags{
				Connection: "192.168.2.10:6790",
				Command:    tc.Command,
				Error:      "",
			}, data.Tags)
		}
	})
}

func TestNoPanicWithNotParsableConnectionString(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	instana.InstrumentSQLDriver(c, "test_driver", sqlDriver{})
	require.Contains(t, sql.Drivers(), "test_driver_with_instana")

	assert.NotPanics(t, func() {
		_, _ = instana.SQLOpen("test_driver",
			"postgres:mysecretpassword@localhost/postgres")
	})
}

func TestProcedureWithCheckerOnStmt(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	var called bool

	driver := &db2DriverMock{
		called: &called,
	}

	instana.InstrumentSQLDriver(c, "test_driver2", driver)
	db, err := instana.SQLOpen("test_driver2", "some datasource")

	assert.NoError(t, err)

	var outValue string
	_, err = db.Exec("CALL SOME_PROCEDURE(?)", sql.Out{Dest: &outValue})

	assert.True(t, called)

	// Here we expect the instrumentation to look for the driver's conn.CheckNamedValue implementation.
	// If there is none, we check stmt.CheckNamedValue, which sqlDriver2 has.
	// If there is none, we return nil from our side, since driver.ErrSkip won't work for CheckNamedValue, as seen here:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.19.1:src/database/sql/driver/driver.go;l=143
	// and here: https://cs.opensource.google/go/go/+/refs/tags/go1.19.1:src/database/sql/driver/driver.go;l=399
	assert.NoError(t, err)
}

func TestProcedureWithNoDefaultChecker(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	driver := pqDriverMock{}

	instana.InstrumentSQLDriver(c, "test_driver3", driver)
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

type sqlDriver struct {
	// Error is a generic error in the SQL execution. It generates spans with errors
	Error error
	// StmtError will give an error when a method from Stmt returns. It does not generate spans at all
	StmtError error
	// PrepareError will give an error when a method from Prepare* returns. It does not generate spans at all
	PrepareError error
}

func (drv sqlDriver) Open(name string) (driver.Conn, error) {
	return sqlConn{
		Error:        drv.Error,
		StmtError:    drv.StmtError,
		PrepareError: drv.PrepareError,
	}, nil
} //nolint:gosimple

type sqlConn struct {
	Error        error
	StmtError    error
	PrepareError error
}

var _ driver.Conn = (*sqlConn)(nil)
var _ driver.ConnPrepareContext = (*sqlConn)(nil)

func (conn sqlConn) Prepare(query string) (driver.Stmt, error) {
	return sqlStmt{Error: conn.Error}, nil
}                                              //nolint:gosimple
func (conn sqlConn) Close() error              { return driver.ErrSkip }
func (conn sqlConn) Begin() (driver.Tx, error) { return nil, driver.ErrSkip }

func (conn sqlConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return sqlStmt{StmtError: conn.StmtError, Error: conn.Error}, conn.PrepareError //nolint:gosimple
}

func (conn sqlConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return sqlResult{}, conn.Error
}

func (conn sqlConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return sqlRows{}, conn.Error
}

type sqlStmt struct {
	Error     error
	StmtError error
}

func (sqlStmt) Close() error                                         { return nil }
func (sqlStmt) NumInput() int                                        { return -1 }
func (stmt sqlStmt) Exec(args []driver.Value) (driver.Result, error) { return sqlResult{}, stmt.Error }
func (stmt sqlStmt) Query(args []driver.Value) (driver.Rows, error)  { return sqlRows{}, stmt.Error }
func (stmt sqlStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return sqlRows{}, stmt.Error
}
func (stmt sqlStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return sqlResult{}, stmt.Error
}

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

type db2DriverMock struct {
	Error  error
	called *bool
}

func (drv *db2DriverMock) Open(name string) (driver.Conn, error) {
	return db2ConnMock{drv.Error, drv.called}, nil
} //nolint:gosimple

type db2ConnMock struct {
	Error  error
	called *bool
}

func (conn db2ConnMock) Prepare(query string) (driver.Stmt, error) {
	return db2StmtMock{conn.Error, conn.called}, nil //nolint:gosimple
}
func (s db2ConnMock) Close() error { return driver.ErrSkip }

func (s db2ConnMock) Begin() (driver.Tx, error) { return nil, driver.ErrSkip }

type db2StmtMock struct {
	Error  error
	called *bool
}

func (db2StmtMock) Close() error  { return nil }
func (db2StmtMock) NumInput() int { return -1 }
func (stmt db2StmtMock) Exec(args []driver.Value) (driver.Result, error) {
	return db2ResultMock{}, stmt.Error
}

func (stmt db2StmtMock) Query(args []driver.Value) (driver.Rows, error) {
	return db2RowsMock{}, stmt.Error
}

func (stmt db2StmtMock) CheckNamedValue(d *driver.NamedValue) error {
	*stmt.called = true
	return nil
}

type db2ResultMock struct{}

func (db2ResultMock) LastInsertId() (int64, error) { return 42, nil }
func (db2ResultMock) RowsAffected() (int64, error) { return 100, nil }

type db2RowsMock struct{}

func (db2RowsMock) Columns() []string              { return []string{"col1", "col2"} }
func (db2RowsMock) Close() error                   { return nil }
func (db2RowsMock) Next(dest []driver.Value) error { return io.EOF }

// Driver use case: driver does not implement NamedValueChecker,arg type checking is internal.
// The idea is to mock pq: https://github.com/lib/pq/blob/8446d16b8935fdf2b5c0fe333538ac395e3e1e4b/encode.go#L31

type pqDriverMock struct{ Error error }

func (drv pqDriverMock) Open(name string) (driver.Conn, error) { return pqConnMock{drv.Error}, nil } //nolint:gosimple

type pqConnMock struct{ Error error }

func (conn pqConnMock) Prepare(query string) (driver.Stmt, error) { return pqStmtMock{conn.Error}, nil } //nolint:gosimple
func (s pqConnMock) Close() error                                 { return driver.ErrSkip }
func (s pqConnMock) Begin() (driver.Tx, error)                    { return nil, driver.ErrSkip }

func (conn pqConnMock) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	var err error

	if _, ok := args[0].Value.(int32); ok {
		err = errors.New("invalid type int32")
	}

	return pqResultMock{}, err
}

func (conn pqConnMock) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return pqRowsMock{}, conn.Error
}

type pqStmtMock struct{ Error error }

func (pqStmtMock) Close() error  { return nil }
func (pqStmtMock) NumInput() int { return -1 }
func (stmt pqStmtMock) Exec(args []driver.Value) (driver.Result, error) {
	return pqResultMock{}, stmt.Error
}
func (stmt pqStmtMock) Query(args []driver.Value) (driver.Rows, error) {
	return pqRowsMock{}, stmt.Error
}

type pqResultMock struct{}

func (pqResultMock) LastInsertId() (int64, error) { return 42, nil }
func (pqResultMock) RowsAffected() (int64, error) { return 100, nil }

type pqRowsMock struct{}

func (pqRowsMock) Columns() []string              { return []string{"col1", "col2"} }
func (pqRowsMock) Close() error                   { return nil }
func (pqRowsMock) Next(dest []driver.Value) error { return io.EOF }
