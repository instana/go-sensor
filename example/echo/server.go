// SPDX-FileCopyrightText: 2026 IBM Corp.
//
// SPDX-License-Identifier: MIT

package main

import (
	"database/sql"
	"log"

	instana "github.com/instana/go-sensor"
	instaechov2 "github.com/instana/go-sensor/instrumentation/instaecho/v2"
	"github.com/labstack/echo/v5"
	_ "modernc.org/sqlite"
)

func main() {
	// Initialize Instana collector
	collector := instana.InitCollector(&instana.Options{
		Service: "echo-v5-example",
	})

	// Set up in-memory SQLite database with Instana instrumentation
	// SQLInstrumentAndOpen is a drop-in replacement for sql.Open() that
	// automatically instruments the driver for distributed tracing
	db, err := instana.SQLInstrumentAndOpen(collector, "sqlite", ":memory:")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create a sample table and insert data
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (name, email) VALUES 
		('Alice', 'alice@example.com'),
		('Bob', 'bob@example.com'),
		('Charlie', 'charlie@example.com')
	`)
	if err != nil {
		log.Fatal("Failed to insert data:", err)
	}

	// Create instrumented Echo v5 instance
	e := instaechov2.New(collector)

	// Define routes - they will be automatically instrumented
	e.GET("/myendpoint", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "pong",
		})
	})

	// Add a new endpoint that demonstrates context propagation to database
	// The trace context flows from the HTTP request through to the database query
	e.GET("/users/:id", func(c *echo.Context) error {
		// Get the context from the request - this contains the trace information
		// from the instrumented Echo handler
		ctx := c.Request().Context()

		// Query the database using the context - this propagates the trace
		// The instrumented SQL driver will automatically create an exit span
		// for this database operation, linking it to the parent HTTP entry span
		var user struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		err := db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = ?", c.Param("id")).
			Scan(&user.ID, &user.Name, &user.Email)

		if err == sql.ErrNoRows {
			return c.JSON(404, map[string]string{
				"error": "User not found",
			})
		}
		if err != nil {
			return c.JSON(500, map[string]string{
				"error": "Database error",
			})
		}

		return c.JSON(200, user)
	})

	// Start server
	log.Fatal(e.Start(":8080"))
}
