// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"database/sql"

	instana "github.com/instana/go-sensor"
)

// This example demonstrates how to instrument an *sql.DB instance created with sql.OpenDB() and driver.Connector
func ExampleWrapSQLConnector() {
	// Here we initialize a new instance of instana.Sensor, however it is STRONGLY recommended
	// to use a single instance throughout your application
	sensor := instana.NewSensor("my-http-client")

	// Instrument the connector. Normally this would be a type provided by the driver library.
	// Here we use a test mock to avoid bringing external dependencies.
	//
	// Note that instana.WrapSQLConnector() requires the connection string to send it later
	// along with database spans.
	connector := instana.WrapSQLConnector(sensor, "driver connection string", &sqlConnector{})

	// Use wrapped connector to initialize the database client. Note that
	db := sql.OpenDB(connector)

	// Inject parent span into the context
	span := sensor.Tracer().StartSpan("entry")
	ctx := instana.ContextWithSpan(context.Background(), span)

	// Query the database, passing the context containing the active span
	db.QueryContext(ctx, "SELECT * FROM users;")

	// SQL queries that are not expected to return a result set are also supported
	db.ExecContext(ctx, "UPDATE users SET last_seen_at = NOW();")
}
