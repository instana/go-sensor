package main

import (
	"time"

	"github.com/instana/golang-sensor"
)

const (
	service = "go-microservice-event"
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
		instana.SendDefaultServiceEvent("Event from the Go sensor sample", "{field: data}", -1, 5000*time.Millisecond)
		time.Sleep(5000 * time.Millisecond)
	}
}
