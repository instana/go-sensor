// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package main

import (
	"time"

	instana "github.com/instana/go-sensor"
)

const (
	service = "games"
)

func main() {
	instana.InitCollector(&instana.Options{
		Service:  service,
		LogLevel: instana.Debug,
		Tracer:   instana.DefaultTracerOptions(),
	})

	go forever()
	select {}
}

func forever() {
	for {
		print("Sending event on service:", service)
		instana.SendServiceEvent("games",
			"Games High Latency", "Games - High latency from East Asia POP.",
			instana.SeverityCritical, 1*time.Second)
		time.Sleep(30 * time.Second)
	}
}
