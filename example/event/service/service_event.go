package main

import (
	"time"

	instana "github.com/instana/go-sensor"
)

const (
	service = "games"
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
		print("Sending event on service:", service)
		instana.SendServiceEvent("games",
			"Games High Latency", "Games - High latency from East Asia POP.",
			instana.SeverityCritical, 1*time.Second)
		time.Sleep(30 * time.Second)
	}
}
