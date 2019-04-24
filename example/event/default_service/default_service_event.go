package main

import (
	"time"

	instana "github.com/instana/go-sensor"
)

const (
	service = "go-microservice-14c"
)

func main() {
	instana.InitSensor(&instana.Options{
		Service:  service,
		LogLevel: instana.Debug})

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
