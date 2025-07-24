// (c) Copyright IBM Corp. 2024

package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"
	"net/http"

	instana "github.com/instana/go-sensor"
	"github.com/mattn/go-sqlite3"
)

func main() {
	mux := http.NewServeMux()

	// Initialize Instana Sensor
	collector := instana.InitCollector(&instana.Options{
		Service: "my-service",
		Tracer:  instana.DefaultTracerOptions(),
	})

	// Instrument the HTTP handler using Instana
	mux.HandleFunc("/adduser", instana.TracingHandlerFunc(collector, "/adduser", handlerHome(collector.LegacySensor())))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Println("error with HTTP server:", err.Error())
		return
	}
}

func handlerHome(sensor *instana.Sensor) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		dsn := "./users.db"

		connector := &CustomConnector{
			driver: &sqlite3.SQLiteDriver{},
			dsn:    dsn,
		}

		// Instrument the sql connector using Instana
		wc := instana.WrapSQLConnector(sensor, "driver connection string", connector)

		// Use the wrapped connector to initialize the database client
		db := sql.OpenDB(wc)
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {
				log.Println("error closing db:", err.Error())
			}
		}(db)

		// IMPORTANT: Since the HTTP handler is instrumented, it is essential to pass the request context for trace
		// propagation.
		ctx := r.Context()

		// Sample queries: Create a table if it doesn't already exist
		createTableSQL := `CREATE TABLE IF NOT EXISTS users (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, name TEXT, age INTEGER);`
		_, err := db.ExecContext(ctx, createTableSQL)
		if err != nil {
			log.Fatal(err)
		}

		// Sample queries: Insert an entry into the table
		name := "John Doe"
		age := 30
		insertUserSQL := `INSERT INTO users (name, age) VALUES (?, ?)`
		_, err = db.ExecContext(ctx, insertUserSQL, name, age)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Entry added to the database successfully!")

		w.Write([]byte("Entry added to the database successfully!"))
	}

}

type CustomConnector struct {
	driver *sqlite3.SQLiteDriver
	dsn    string
}

// Connect establishes a new connection to the database
func (c *CustomConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.driver.Open(c.dsn)
}

// Driver returns the underlying driver of the connector
func (c *CustomConnector) Driver() driver.Driver {
	return c.driver
}
