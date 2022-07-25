// (c) Copyright IBM Corp. 2022

//go:build go1.16
// +build go1.16

package instaredigo_test

import (
	"fmt"
	"os"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaredigo"
)

func Example_batchCommands() {
	// Create a sensor for instana instrumentation
	sensor := instana.NewSensor("mysensor")

	//Create an InstaRedigo connection
	conn, err := instaredigo.Dial(sensor, "tcp", ":7001")
	if err != nil {
		os.Exit(1)
	}
	defer conn.Close()

	// Send a batch of commands  using the new connection
	err = conn.Send("MULTI")
	err = conn.Send("INCR", "foo")
	err = conn.Send("INCR", "bar")
	reply, err := conn.Do("EXEC")
	if err != nil {
		fmt.Println("Error while sending command. Details: ", err.Error())
	}
	fmt.Println("Response received: ", reply)
}
