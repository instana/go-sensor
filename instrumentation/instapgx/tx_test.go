// (c) Copyright IBM Corp. 2022

//go:build integration
// +build integration

package instapgx_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	instana "github.com/instana/go-sensor"

	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestBeginCommit(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)

	testTx(t, ctx, tx, recorder)
}

func TestBeginTxCommit(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	assert.NoError(t, err)

	testTx(t, ctx, tx, recorder)
}

func TestTxBeginFunc(t *testing.T) {
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

	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)

	err = tx.BeginFunc(ctx, func(tx pgx.Tx) error {
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

	assert.Equal(t, 6, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 6)
}

func TestRollback(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	err = tx.Rollback(ctx)
	assert.NoError(t, err)

	assert.Equal(t, 2, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 2)
}

func TestCommit(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, conn := prepare(t)

	tx, err := conn.Begin(ctx)
	assert.NoError(t, err)
	err = tx.Commit(ctx)
	assert.NoError(t, err)

	assert.Equal(t, 2, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 2)
}

func TestBeginTxFunc(t *testing.T) {
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

	err = conn.BeginTxFunc(ctx, pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.Deferrable,
	}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, fmt.Sprintf("insert into %s(id) values (1)", uniqString))
		assert.NoError(t, err)
		return nil
	})
	assert.NoError(t, err)

	assert.Equal(t, 4, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 4)
}

func testTx(t *testing.T, ctx context.Context, tx pgx.Tx, recorder *instana.Recorder) {
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
			rows, err := tx.Query(ctx, sqlStmt)
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

			row := tx.QueryRow(ctx, sqlStmt)
			assert.NoError(t, err, "when exec:", sqlStmt)

			var id int
			var url, name string

			err = row.Scan(&id, &url, &name)
			assert.NoError(t, err, "when scan")

			assert.NotEmpty(t, id)
			assert.NotEmpty(t, url)
			assert.NotEmpty(t, name)

		} else {
			commandTag, err := tx.Exec(ctx, sqlStmt)
			assert.NoError(t, err, "when exec:", sqlStmt)
			assert.NotEmpty(t, commandTag)
		}
	}

	err := tx.Commit(ctx)
	assert.NoError(t, err)

	assert.Equal(t, 8, recorder.QueuedSpansCount())
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 8)
}
