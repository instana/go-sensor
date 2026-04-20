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
	fmt.Println("=== Instana Metrics Transmission Interval - Code Configuration ===")
	fmt.Println()

	// Example 1: Configure metrics transmission interval via code
	fmt.Println("Example 1: Setting custom interval to 3000ms (3 seconds)")

	opts := &instana.Options{
		Service: "metrics-interval-code-example-2",
		Metrics: instana.MetricsOptions{
			TransmissionDelay: 3000, // 3000 milliseconds = 3 seconds
		},
	}

	col := instana.InitCollector(opts)

	fmt.Println("✓ Instana collector initialized")
	fmt.Println("✓ Metrics will be transmitted every 3000ms")
	fmt.Println()

	// Example 2: Different configurations
	fmt.Println("Other configuration examples:")
	fmt.Println()

	// Fast interval for high-frequency monitoring
	fmt.Println("  Fast interval (500ms):")
	fmt.Println("    Metrics: instana.MetricsOptions{")
	fmt.Println("      TransmissionDelay: 500,")
	fmt.Println("    }")
	fmt.Println()

	// Slow interval for resource-constrained environments
	fmt.Println("  Slow interval (5000ms - maximum):")
	fmt.Println("    Metrics: instana.MetricsOptions{")
	fmt.Println("      TransmissionDelay: 5000,")
	fmt.Println("    }")
	fmt.Println()

	// Default behavior
	fmt.Println("  Default interval (1000ms):")
	fmt.Println("    Metrics: instana.MetricsOptions{")
	fmt.Println("      TransmissionDelay: 0, // or omit the field")
	fmt.Println("    }")
	fmt.Println()

	fmt.Println("Configuration rules:")
	fmt.Println("  - Valid range: 1ms to 5000ms")
	fmt.Println("  - Values above 5000ms are automatically capped at 5000ms")
	fmt.Println("  - Zero or negative values use default 1000ms")
	fmt.Println("  - Environment variable INSTANA_METRICS_TRANSMISSION_DELAY takes precedence")
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
