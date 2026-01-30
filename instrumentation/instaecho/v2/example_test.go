// SPDX-FileCopyrightText: 2026 IBM Corp.
// SPDX-FileCopyrightText: 2026 Instana Inc.
//
// SPDX-License-Identifier: MIT

//go:build go1.18
// +build go1.18

package instaechov2_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	instana "github.com/instana/go-sensor"
	instaechov2 "github.com/instana/go-sensor/instrumentation/instaecho/v2"
	"github.com/labstack/echo/v5"
)

// testAgentClient is a mock agent client for examples
type testAgentClient struct{}

func (testAgentClient) Ready() bool                                       { return true }
func (testAgentClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (testAgentClient) SendEvent(event *instana.EventData) error          { return nil }
func (testAgentClient) SendSpans(spans []instana.Span) error              { return nil }
func (testAgentClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (testAgentClient) Flush(context.Context) error                       { return nil }

// Example demonstrates how to instrument an Echo v5 HTTP server with Instana tracing.
func Example() {
	// Initialize Instana collector
	collector := instana.InitCollector(&instana.Options{
		Service:     "echo-v5-service",
		AgentClient: testAgentClient{},
	})
	defer instana.ShutdownCollector()

	// Create an instrumented Echo v5 instance
	e := instaechov2.New(collector)

	// Define a route - it will be automatically instrumented
	e.GET("/hello/:name", func(c *echo.Context) error {
		name := c.Param("name")
		return c.String(http.StatusOK, fmt.Sprintf("Hello, %s!", name))
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/hello/World", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Print the response
	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())

	// Output:
	// 200
	// Hello, World!
}

// Example_middleware demonstrates how to use the Instana middleware with an existing Echo v5 instance.
func Example_middleware() {
	// Initialize Instana collector
	collector := instana.InitCollector(&instana.Options{
		Service:     "echo-v5-custom",
		AgentClient: testAgentClient{},
	})
	defer instana.ShutdownCollector()

	// Create a standard Echo v5 instance
	e := echo.New()

	// Add Instana middleware
	e.Use(instaechov2.Middleware(collector))

	// Define routes
	e.GET("/status", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Print the response
	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())

	// Output:
	// 200
	// OK
}
