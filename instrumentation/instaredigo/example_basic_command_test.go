// (c) Copyright IBM Corp. 2022

//go:build go1.16
// +build go1.16

package instaredigo_test

import (
	"context"
	"fmt"
	"os"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaredigo"
)

func Example_basicCommand() {
	// Create a sensor for instana instrumentation
	sensor := instana.NewSensor("mysensor")

	// Create an InstaRedigo connection
	conn, err := instaredigo.Dial(sensor, "tcp", ":7001")
	if err != nil {
		os.Exit(1)
	}
	defer conn.Close()

	// Send a command using the new connection
	ctx := context.Background()
	reply, err := conn.Do("SET", "greetings", "helloworld", ctx)
	if err != nil {
		fmt.Println("Error while sending command. Details: ", err.Error())
	}
	fmt.Println("Response received: ", fmt.Sprintf("%s", reply))
}
