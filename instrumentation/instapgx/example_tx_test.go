// (c) Copyright IBM Corp. 2022

package instapgx_test

import (
	"context"
	"log"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instapgx"
	"github.com/jackc/pgx/v4"
)

func ExampleTx_Prepare() {
	databaseUrl := "postgres://postgres:mysecretpassword@localhost/postgres"

	sensor := instana.NewSensor("pgx-example-prepare")
	conf, err := pgx.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ctx := context.Background()
	conn, err := instapgx.ConnectConfig(ctx, sensor, conf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer tx.Commit(ctx)

	_, err = tx.Prepare(ctx, "mystatement", "select name, statement from pg_prepared_statements")
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleTx_Begin() {
	databaseUrl := "postgres://postgres:mysecretpassword@localhost/postgres"

	sensor := instana.NewSensor("pgx-example-begin")
	conf, err := pgx.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ctx := context.Background()
	conn, err := instapgx.ConnectConfig(ctx, sensor, conf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ptx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer ptx.Commit(ctx)

	tx, err := ptx.Begin(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		log.Fatalln(err.Error())
	}
}

func ExampleTx_BeginFunc() {
	databaseUrl := "postgres://postgres:mysecretpassword@localhost/postgres"

	sensor := instana.NewSensor("pgx-example-begin-func")
	conf, err := pgx.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ctx := context.Background()
	conn, err := instapgx.ConnectConfig(ctx, sensor, conf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer tx.Commit(ctx)

	err = tx.BeginFunc(ctx, func(tx pgx.Tx) error {
		var a, b int
		_, err = tx.QueryFunc(
			ctx,
			"select n, n * 2 from generate_series(1, $1) n",
			[]interface{}{3},
			[]interface{}{&a, &b},
			func(pgx.QueryFuncRow) error {
				return nil
			},
		)

		return err
	})

	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleTx_Exec() {
	databaseUrl := "postgres://postgres:mysecretpassword@localhost/postgres"

	sensor := instana.NewSensor("pgx-example-exec")
	conf, err := pgx.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ctx := context.Background()
	conn, err := instapgx.ConnectConfig(ctx, sensor, conf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer tx.Commit(ctx)

	_, err = tx.Exec(ctx, ";")
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleTx_Query() {
	databaseUrl := "postgres://postgres:mysecretpassword@localhost/postgres"

	sensor := instana.NewSensor("pgx-example-query")
	conf, err := pgx.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ctx := context.Background()
	conn, err := instapgx.ConnectConfig(ctx, sensor, conf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer tx.Commit(ctx)

	row := tx.QueryRow(ctx, "select name, statement from pg_prepared_statements")

	var name, stmt string
	err = row.Scan(&name, &stmt)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleTx_QueryFunc() {
	databaseUrl := "postgres://postgres:mysecretpassword@localhost/postgres"

	sensor := instana.NewSensor("pgx-example-query-func")
	conf, err := pgx.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ctx := context.Background()
	conn, err := instapgx.ConnectConfig(ctx, sensor, conf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer tx.Commit(ctx)

	var a, b int
	_, err = tx.QueryFunc(
		ctx,
		"select n, n * 2 from generate_series(1, $1) n",
		[]interface{}{3},
		[]interface{}{&a, &b},
		func(pgx.QueryFuncRow) error {
			return nil
		},
	)

	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleTx_SendBatch() {
	databaseUrl := "postgres://postgres:mysecretpassword@localhost/postgres"

	sensor := instana.NewSensor("pgx-example-send-batch")
	conf, err := pgx.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ctx := context.Background()
	conn, err := instapgx.ConnectConfig(ctx, sensor, conf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer tx.Commit(ctx)

	instapgx.EnableDetailedBatchMode()

	b := &pgx.Batch{}
	b.Queue("select name, statement from pg_prepared_statements")
	b.Queue("select n, n * 2 from generate_series(1, $1) n", 1)

	br := tx.SendBatch(ctx, b)

	_, err = br.Query()
	if err != nil {
		log.Fatalln(err.Error())
	}

	_, err = br.Exec()
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = br.Close()
	if err != nil {
		log.Fatalln(err.Error())
	}
}
