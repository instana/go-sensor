// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package main

import (
	"time"

	instana "github.com/instana/go-sensor"
)

const (
	service = "go-microservice-14c"
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
		instana.SendDefaultServiceEvent(
			"Service Restart", "This service has been restarted with change e3b926d by @pglombardo",
			instana.SeverityChange, 5*time.Second)
		time.Sleep(30 * time.Second)
	}
}
