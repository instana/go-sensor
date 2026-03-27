// SPDX-FileCopyrightText: 2026 IBM Corp.
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instafiber/v2"
	"github.com/instana/go-sensor/instrumentation/instagorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Configuration constants
const (
	serverPort        = ":3000"
	databasePath      = "data.db"
	serviceName       = "Fiber Example App"
	agentReadyTimeout = 30 * time.Second
)

var collectableHeaders = []string{"Host", "Connection"}

// Student represents a student record in the database
type Student struct {
	gorm.Model
	Name       string
	RollNumber uint
}

// waitForInstanaAgent blocks until the Instana agent is ready or timeout occurs.
// It returns an error if the agent doesn't become ready within the specified timeout.
func waitForInstanaAgent(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Instana agent to become ready")
		case <-ticker.C:
			if instana.Ready() {
				return nil
			}
		}
	}
}

// initializeInstanaCollector creates and configures the Instana collector
// with the specified service name and HTTP headers to collect.
func initializeInstanaCollector() instana.TracerLogger {
	opts := instana.DefaultTracerOptions()
	opts.CollectableHTTPHeaders = collectableHeaders

	return instana.InitCollector(&instana.Options{
		Service:           serviceName,
		EnableAutoProfile: true,
		Tracer:            opts,
	})
}

// initializeDatabase sets up the database connection, instruments it with Instana,
// and runs the necessary migrations. Returns an error if any step fails.
func initializeDatabase(collector instana.TracerLogger) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	instagorm.Instrument(db, collector, databasePath)

	if err := db.AutoMigrate(&Student{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return db, nil
}

// createStudentHandler returns a Fiber handler that creates a new student record
// in the database and returns a greeting message.
func createStudentHandler(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Set the database context to the request context for tracing
		db.Statement.Context = c.Context()

		// Create a new student record
		db.Create(&Student{Name: "Alex", RollNumber: 32})
		fmt.Println("Student added to DB!")

		// Send a string response to the client
		return c.SendString("Student added to DB!")
	}
}

func main() {
	// Initialize Instana collector
	collector := initializeInstanaCollector()

	// Wait for Instana agent to be ready
	if err := waitForInstanaAgent(agentReadyTimeout); err != nil {
		log.Fatalf("Instana agent initialization failed: %v", err)
	}
	fmt.Println("Instana agent ready")

	// Initialize database with migrations
	db, err := initializeDatabase(collector)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// Setup Fiber application and routes
	app := fiber.New()
	app.Get("/create", instafiber.TraceHandler(collector, "create", "/create", createStudentHandler(db)))

	// Start the HTTP server
	fmt.Printf("Starting server on %s\n", serverPort)
	log.Fatal(app.Listen(serverPort))
}
