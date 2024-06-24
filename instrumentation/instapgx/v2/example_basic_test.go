// (c) Copyright IBM Corp. 2024

package instapgx_test

import (
	"context"
	"fmt"
	"os"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instapgx/v2"
	"github.com/jackc/pgx/v5"
)

func Example_basicUsage() {
	urlExample := "postgres://postgres:mysecretpassword@localhost:5432/postgres"
	cfg, err := pgx.ParseConfig(urlExample)

	// Initialising Instana Sensor
	sensor := instana.NewSensor("pgx-v5-service")
	// Assigning the Instana tracer to the cfg.Tracer interface
	cfg.Tracer = instapgx.InstanaTracer(cfg, sensor)

	// Use the cfg in the normal way to create a connection and use it
	ctx := context.Background()
	var conn *pgx.Conn
	conn, err = pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		fmt.Printf("unable to connect to database: %v\n\n", err)
		os.Exit(1)
	} else {
		fmt.Println("connection successful")
	}
	defer func() {
		err := conn.Close(ctx)
		if err != nil {
			fmt.Printf("unable to close connection: %v\n", err)
		}
	}()

	var val string
	query := "<valid-query>"
	err = conn.QueryRow(ctx, query).Scan(&val)
	if err != nil {
		fmt.Printf("queryRow failed: %v\n\n", err)
		os.Exit(1)
	}
}
