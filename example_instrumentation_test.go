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
	sensor := instana.NewSensor("my-http-server")

	http.HandleFunc("/", instana.TracingHandlerFunc(sensor, "root", func(w http.ResponseWriter, req *http.Request) {
		// handler code
	}))
}

// This example demonstrates how to instrument an HTTP client with Instana
func ExampleRoundTripper() {
	// Here we initialize a new instance of instana.Sensor, however it is STRONGLY recommended
	// to use a single instance throughout your application
	sensor := instana.NewSensor("my-http-client")
	span := sensor.Tracer().StartSpan("entry")

	// http.DefaultTransport is used as a default RoundTripper, however you can provide
	// your own implementation
	client := &http.Client{
		Transport: instana.RoundTripper(sensor, nil),
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
	sensor := instana.NewSensor("my-http-client")

	// Instrument the driver. Normally this would be a type provided by the driver library, e.g.
	// pq.Driver{} or mysql.Driver{}, but here we use a test mock to avoid bringing external dependencies
	instana.InstrumentSQLDriver(sensor, "your_db_driver", sqlDriver{})

	// Replace sql.Open() with instana.SQLOpen()
	db, _ := instana.SQLOpen("your_db_driver", "driver connection string")

	// Inject parent span into the context
	span := sensor.Tracer().StartSpan("entry")
	ctx := instana.ContextWithSpan(context.Background(), span)

	// Query the database, passing the context containing the active span
	db.QueryContext(ctx, "SELECT * FROM users;")

	// SQL queries that are not expected to return a result set are also supported
	db.ExecContext(ctx, "UPDATE users SET last_seen_at = NOW();")
}
