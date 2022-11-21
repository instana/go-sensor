// (c) Copyright IBM Corp. 2022

//go:build integration
// +build integration

package instapgx_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instapgx"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
)

func TestConnect(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)
	defer instana.ShutdownSensor()

	ctx := context.Background()
	conn, err := instapgx.Connect(ctx, sensor, databaseUrl)

	assert.NoError(t, err)
	assert.IsType(t, &instapgx.Conn{}, conn)
}

func TestConnectConfig(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)
	defer instana.ShutdownSensor()

	conf, err := pgx.ParseConfig(databaseUrl)
	assert.NoError(t, err)
	ctx := context.Background()
	conn, err := instapgx.ConnectConfig(ctx, sensor, conf)

	assert.NoError(t, err)
	assert.IsType(t, &instapgx.Conn{}, conn)
}

func TestExecAndQueryWithoutParameters(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	uniqString := randStringBytes(10)
	sqls := []string{
		fmt.Sprintf("DROP TABLE IF EXISTS %s;", uniqString),
		fmt.Sprintf(`CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			url VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description VARCHAR (255),
			last_update DATE
		);`, uniqString),
		fmt.Sprintf(`INSERT INTO %s (url, name) VALUES('https://instana.com','Instana website');`, uniqString),
		fmt.Sprintf(`INSERT INTO %s (url, name) VALUES('https://www.instana.com/blog/','Instana blog');`, uniqString),
		fmt.Sprintf(`SELECT id, url, name FROM %s`, uniqString),
	}

	for _, sqlStmt := range sqls {
		if strings.HasPrefix(sqlStmt, "SELECT") {
			rows, err := conn.Query(ctx, sqlStmt)
			assert.NoError(t, err, "when exec:", sqlStmt)
			for rows.Next() {
				var id int
				var url, name string

				err := rows.Scan(&id, &url, &name)
				assert.NoError(t, err, "when scan")

				assert.NotEmpty(t, id)
				assert.NotEmpty(t, url)
				assert.NotEmpty(t, name)
			}
			rows.Close()

			row := conn.QueryRow(ctx, sqlStmt)
			assert.NoError(t, err, "when exec:", sqlStmt)

			var id int
			var url, name string

			err = row.Scan(&id, &url, &name)
			assert.NoError(t, err, "when scan")

			assert.NotEmpty(t, id)
			assert.NotEmpty(t, url)
			assert.NotEmpty(t, name)

		} else {
			commandTag, err := conn.Exec(ctx, sqlStmt)
			assert.NoError(t, err, "when exec:", sqlStmt)
			assert.NotEmpty(t, commandTag)
		}
	}

	assert.Equal(t, 6, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 6)
}

func TestExecWithParameters(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	uniqString := randStringBytes(10)
	sqls := []string{
		fmt.Sprintf("DROP TABLE IF EXISTS %s;", uniqString),
		fmt.Sprintf(`CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			url VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description VARCHAR (255),
			last_update DATE
		);`, uniqString),
		fmt.Sprintf(`INSERT INTO %s (url, name) VALUES($1, $2);`, uniqString),
		fmt.Sprintf(`INSERT INTO %s (url, name) VALUES($1, $2);`, uniqString),
		fmt.Sprintf(`SELECT id, url, name FROM %s WHERE '1'=$1`, uniqString),
	}
	params := [][]string{
		{},
		{},
		{`https://instana.com`, `Instana website`},
		{`https://www.instana.com/blog/`, `Instana blog`},
		{`1`},
	}

	for n, sqlStmt := range sqls {
		if strings.HasPrefix(sqlStmt, "SELECT") {
			rows, err := conn.Query(ctx, sqlStmt, params[n][0])
			assert.NoError(t, err, "when exec:", sqlStmt, params[n][0])
			for rows.Next() {
				var id int
				var url, name string
				err := rows.Scan(&id, &url, &name)
				assert.NoError(t, err)

				assert.NotEmpty(t, id)
				assert.NotEmpty(t, url)
				assert.NotEmpty(t, name)
			}
			rows.Close()

			row := conn.QueryRow(ctx, sqlStmt, params[n][0])
			assert.NoError(t, err, "when exec:", sqlStmt, params[n][0])

			var id int
			var url, name string

			err = row.Scan(&id, &url, &name)
			assert.NoError(t, err)

			assert.NotEmpty(t, id)
			assert.NotEmpty(t, url)
			assert.NotEmpty(t, name)

		} else {
			var commandTag pgconn.CommandTag
			var err error
			if len(params[n]) == 2 {
				commandTag, err = conn.Exec(ctx, sqlStmt, params[n][0], params[n][1])
			} else {
				commandTag, err = conn.Exec(ctx, sqlStmt)
			}

			assert.NoError(t, err, "when exec:", sqlStmt)
			assert.NotEmpty(t, commandTag)
		}
	}

	assert.Equal(t, 6, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 6)
}

func TestQueryFunc(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, _, conn := prepare(t)

	var a, b int
	_, err := conn.QueryFunc(
		context.Background(),
		"select n, n * 2 from generate_series(1, $1) n",
		[]interface{}{3},
		[]interface{}{&a, &b},
		func(pgx.QueryFuncRow) error {
			return nil
		},
	)

	assert.NoError(t, err)

	assert.Equal(t, 1, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 1)
}

func TestQueryFuncError(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, _, conn := prepare(t)

	var a, b int
	_, err := conn.QueryFunc(
		context.Background(),
		"select n, n * 2 from BAD_FUNCTION(1, $1) n",
		[]interface{}{3},
		[]interface{}{&a, &b},
		func(pgx.QueryFuncRow) error {
			return nil
		},
	)

	assert.Error(t, err)

	assert.Equal(t, 2, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 2)

	d, err := json.Marshal(spans[1].Data)
	assert.NoError(t, err)
	assert.Contains(t, string(d), `ERROR: function bad_function(integer, unknown) does not exist (SQLSTATE 42883)`)
}

func TestSendBatchError(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	b := &pgx.Batch{}
	b.Queue("select n, n * 2 from BAD_FUNCTION(1, $1) n")
	br := conn.SendBatch(ctx, b)

	_, err := br.Query()
	assert.Error(t, err)

	assert.Equal(t, 3, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 3)

	d, err := json.Marshal(spans[2].Data)
	assert.NoError(t, err)
	assert.Contains(t, string(d), `ERROR: function bad_function(integer, unknown) does not exist (SQLSTATE 42883)`)
}

func TestSendBatchQuery(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	b := &pgx.Batch{}
	b.Queue("select n, n * 2 from generate_series(1, $1) n", 1)
	br := conn.SendBatch(ctx, b)

	_, err := br.Query()
	assert.NoError(t, err)

	assert.Equal(t, 2, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 2)
}

func TestSendBatchExec(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	b := &pgx.Batch{}
	b.Queue("select n, n * 2 from generate_series(1, $1) n", 1)
	br := conn.SendBatch(ctx, b)

	_, err := br.Exec()
	assert.NoError(t, err)

	assert.Equal(t, 2, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 2)
}

func TestSendBatchQueryRow(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	b := &pgx.Batch{}
	b.Queue("select n, n * 2 from generate_series(1, $1) n", 1)
	br := conn.SendBatch(ctx, b)

	r := br.QueryRow()

	var v1, v2 int
	err := r.Scan(&v1, &v2)
	assert.NoError(t, err)

	assert.Equal(t, 2, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 2)
}

func TestSendBatchQueryRowError(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	b := &pgx.Batch{}
	b.Queue("select n, n * 2 from generate_series(1, $1) n", 1)
	br := conn.SendBatch(ctx, b)

	r := br.QueryRow()

	var v1, v2 string
	err := r.Scan(&v1, &v2)
	assert.Error(t, err)

	assert.Equal(t, 3, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 3)
}

func TestSendBatchQueryRowScanMultipleTimes(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	b := &pgx.Batch{}
	b.Queue("select n, n * 2 from generate_series(1, $1) n", 1)
	br := conn.SendBatch(ctx, b)

	r := br.QueryRow()

	wg := sync.WaitGroup{}

	i := 10
	wg.Add(10)

	for i > 0 {
		go func() {
			var v1, v2 int
			_ = r.Scan(&v1, &v2)
			wg.Done()
		}()

		i--
	}

	wg.Wait()

	assert.Equal(t, 2, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 2)
}

func TestSendBatchQueryFunc(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	b := &pgx.Batch{}
	b.Queue("select n, n * 2 from generate_series(1, $1) n", 1)
	br := conn.SendBatch(ctx, b)

	var v1, v2 int
	_, err := br.QueryFunc([]interface{}{&v1, &v2},
		func(pgx.QueryFuncRow) error {
			return nil
		},
	)

	assert.NoError(t, err)

	assert.Equal(t, 2, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 2)
}

func TestSendBatchExecTwice(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	b := &pgx.Batch{}
	b.Queue("select n, n * 2 from generate_series(1, $1) n", 1)
	br := conn.SendBatch(ctx, b)

	_, err := br.Exec()
	assert.NoError(t, err)
	_, err = br.Exec()
	assert.Error(t, err)

	assert.Equal(t, 4, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 4)
}

func TestSendBatchClose(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	b := &pgx.Batch{}
	b.Queue("select n, n * 2 from generate_series(1, $1) n", 1)
	br := conn.SendBatch(ctx, b)

	err := br.Close()
	assert.NoError(t, err)

	assert.Equal(t, 1, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 1)
}

func TestCopyFrom(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, _, conn := prepare(t)

	_, err := conn.Exec(context.Background(), `create temporary table foo(a int2, b int4, c int8, d varchar, e text, f date, g timestamptz)`)
	assert.NoError(t, err)

	tzedTime := time.Date(2010, 2, 3, 4, 5, 6, 0, time.Local)

	inputRows := [][]interface{}{
		{int16(0), int32(1), int64(2), "abc", "efg", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), tzedTime},
		{nil, nil, nil, nil, nil, nil, nil},
	}

	copyCount, err := conn.CopyFrom(context.Background(), pgx.Identifier{"foo"}, []string{"a", "b", "c", "d", "e", "f", "g"}, pgx.CopyFromRows(inputRows))
	assert.NoError(t, err)
	assert.EqualValues(t, len(inputRows), copyCount)

	rows, err := conn.Query(context.Background(), "select * from foo")
	assert.NoError(t, err)

	var outputRows [][]interface{}
	for rows.Next() {
		row, err := rows.Values()
		if err != nil {
			t.Errorf("Unexpected error for rows.Values(): %v", err)
		}
		outputRows = append(outputRows, row)
	}

	assert.NoError(t, rows.Err())
	assert.Equal(t, inputRows, outputRows)

	assert.Equal(t, 3, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 3)
}

func TestPrepare(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	_, err := conn.Exec(context.Background(), `create temporary table foo(a int2, b int4, c int8, d varchar, e text, f date, g timestamptz)`)
	assert.NoError(t, err)

	_, err = conn.Prepare(ctx, "test", "select * from foo;")
	assert.NoError(t, err)

	assert.Equal(t, 2, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 2)
}

func TestBeginFunc(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	uniqString := randStringBytes(10)
	createSql := fmt.Sprintf(`
    create temporary table %s(
      id integer,
      unique (id)
    );
  `, uniqString)

	_, err := conn.Exec(ctx, createSql)
	assert.NoError(t, err)

	err = conn.BeginFunc(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, fmt.Sprintf("insert into %s(id) values (1)", uniqString))
		assert.NoError(t, err)

		var a, b int
		_, err = tx.QueryFunc(
			context.Background(),
			"select n, n * 2 from generate_series(1, $1) n",
			[]interface{}{3},
			[]interface{}{&a, &b},
			func(pgx.QueryFuncRow) error {
				return nil
			},
		)
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)

	assert.Equal(t, 5, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 5)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
