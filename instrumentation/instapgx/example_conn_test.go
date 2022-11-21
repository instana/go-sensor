// (c) Copyright IBM Corp. 2022

package instapgx_test

import (
	"context"
	"log"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instapgx"
	"github.com/jackc/pgx/v4"
)

func ExampleConn_Prepare() {
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

	_, err = conn.Prepare(ctx, "mystatement", "select name, statement from pg_prepared_statements")
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleConn_Begin() {
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

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		log.Fatalln(err.Error())
	}
}

func ExampleConn_BeginFunc() {
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

	err = conn.BeginFunc(ctx, func(tx pgx.Tx) error {
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

func ExampleConn_Ping() {
	databaseUrl := "postgres://postgres:mysecretpassword@localhost/postgres"

	sensor := instana.NewSensor("pgx-example-ping")
	conf, err := pgx.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ctx := context.Background()
	conn, err := instapgx.ConnectConfig(ctx, sensor, conf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = conn.Ping(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleConn_Exec() {
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

	_, err = conn.Exec(ctx, "VACUUM (VERBOSE, ANALYZE)")
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleConn_Query() {
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

	row := conn.QueryRow(ctx, "select name, statement from pg_prepared_statements")

	var name, stmt string
	err = row.Scan(&name, &stmt)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleConn_QueryFunc() {
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

	var a, b int
	_, err = conn.QueryFunc(
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

func ExampleConn_SendBatch() {
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

	instapgx.EnableDetailedBatchMode()

	b := &pgx.Batch{}
	b.Queue("select name, statement from pg_prepared_statements")
	b.Queue("select n, n * 2 from generate_series(1, $1) n", 1)

	br := conn.SendBatch(ctx, b)

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
