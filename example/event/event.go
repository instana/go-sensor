package main

import (
	"time"

	"github.com/instana/golang-sensor"
)

const (
	service = "golang-event"
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
		instana.SendDefaultServiceEvent("Event from the Golang sample", "{field: data}", -1)
		time.Sleep(5000 * time.Millisecond)
	}
}
