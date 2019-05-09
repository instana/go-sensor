package main

import (
	"time"

	instana "github.com/instana/go-sensor"
)

func main() {
	go forever()
	select {}
}

func forever() {
	for {
		instana.SendServiceEvent("go-microservice-14c",
			"Go Microservice Online", "♪┏(°.°)┛┗(°.°)┓┗(°.°)┛┏(°.°)┓ ♪", instana.SeverityChange, 1*time.Second)
		time.Sleep(30000 * time.Millisecond)
	}
}
