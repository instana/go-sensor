// SPDX-FileCopyrightText: 2026 IBM Corp.
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	instana "github.com/instana/go-sensor"
)

func main() {
	// Example: Configure metrics transmission interval via environment variable
	// Set INSTANA_METRICS_TRANSMISSION_DELAY before starting the application

	// For demonstration, we'll set it programmatically here
	// In production, set this via your deployment configuration
	os.Setenv("INSTANA_METRICS_TRANSMISSION_DELAY", "2000")

	fmt.Println("=== Instana Metrics Transmission Interval - ENV Configuration ===")
	fmt.Println()
	fmt.Println("Environment variable INSTANA_METRICS_TRANSMISSION_DELAY=2000")
	fmt.Println("This will configure metrics to be transmitted every 2000ms (2 seconds)")
	fmt.Println()

	// Initialize the Instana collector with default options
	// The environment variable will be automatically applied
	col := instana.InitCollector(&instana.Options{
		Service: "metrics-interval-env-example",
	})

	fmt.Println("✓ Instana collector initialized")
	fmt.Println("✓ Metrics will be transmitted every 2000ms")
	fmt.Println()
	fmt.Println("Valid values:")
	fmt.Println("  - Minimum: 1ms")
	fmt.Println("  - Maximum: 5000ms (values above will be capped)")
	fmt.Println("  - Default: 1000ms (if not specified or invalid)")
	fmt.Println()
	fmt.Println("Invalid values (non-numeric, negative, zero) will fall back to default 1000ms")
	fmt.Println()

	// Simulate application running
	go func() {
		http.HandleFunc("/endpoint", instana.TracingHandlerFunc(col, "/endpoint", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		log.Fatal(http.ListenAndServe(":7070", nil))
	}()

	go func() {
		ticker := time.NewTicker(5 * time.Second)

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		for range ticker.C {
			url := "http://localhost:7070/endpoint"
			// Create request
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				fmt.Println("Error creating request:", err)
				return
			}
			// Send request
			_, err = client.Do(req)
			if err != nil {
				fmt.Println("Error making request:", err)
				return
			}
		}
	}()

	fmt.Println("Please go to the Instana UI to see metrics")
	fmt.Println("Application running... (press Ctrl+C to exit)")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println("Application stopped.")
}
