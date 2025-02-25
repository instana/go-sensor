// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"net/http"

	instana "github.com/instana/go-sensor"
)

// This example demonstrates how to instrument an HTTP handler with Instana and register it
// in http.DefaultServeMux
func ExampleTracingHandlerFunc() {
	// Here we initialize a new instance of instana.Sensor, however it is STRONGLY recommended
	// to use a single instance throughout your application
	c := instana.InitCollector(&instana.Options{
		Service: "my-http-server",
	})
	defer instana.ShutdownCollector()

	http.HandleFunc("/", instana.TracingNamedHandlerFunc(c, "root", "/", func(w http.ResponseWriter, req *http.Request) {
		// handler code
	}))
}

// This example demonstrates how to instrument an HTTP client with Instana
func ExampleRoundTripper() {
	// Here we initialize a new instance of instana.Sensor, however it is STRONGLY recommended
	// to use a single instance throughout your application
	c := instana.InitCollector(&instana.Options{
		Service: "my-http-server",
	})
	defer instana.ShutdownCollector()

	span := c.Tracer().StartSpan("entry")

	// Make sure to finish the span so it can be properly recorded in the backend
	defer span.Finish()

	// http.DefaultTransport is used as a default RoundTripper, however you can provide
	// your own implementation
	client := &http.Client{
		Transport: instana.RoundTripper(c, nil),
	}

	// Inject parent span into the request context
	ctx := instana.ContextWithSpan(context.Background(), span)
	req, _ := http.NewRequest("GET", "https://www.instana.com", nil)

	// Execute request as usual
	client.Do(req.WithContext(ctx))
}

// This example demonstrates how to instrument an *sql.DB instance created with sql.Open()
func ExampleSQLOpen() {
	// Here we initialize a new instance of instana.Sensor, however it is STRONGLY recommended
	// to use a single instance throughout your application
	c := instana.InitCollector(&instana.Options{
		Service: "my-http-server",
	})
	defer instana.ShutdownCollector()

	// Instrument the driver. Normally this would be a type provided by the driver library, e.g.
	// pq.Driver{} or mysql.Driver{}, but here we use a test mock to avoid bringing external dependencies
	instana.InstrumentSQLDriver(c, "your_db_driver", sqlDriver{})

	// Replace sql.Open() with instana.SQLOpen()
	db, _ := instana.SQLOpen("your_db_driver", "driver connection string")

	// Inject parent span into the context
	span := c.Tracer().StartSpan("entry")
	ctx := instana.ContextWithSpan(context.Background(), span)

	// Query the database, passing the context containing the active span
	db.QueryContext(ctx, "SELECT * FROM users;")

	// SQL queries that are not expected to return a result set are also supported
	db.ExecContext(ctx, "UPDATE users SET last_seen_at = NOW();")
}
