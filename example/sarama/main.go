// (c) Copyright IBM Corp. 2023

package main

import "time"

func main() {

	createTopic()
	ch := make(chan bool)
	go consume(ch)
	produce()

	// The message was received by the consumer, so we can move on.
	<-ch

	// Give the Go tracer some time to send data to the agent.
	// This step is not needed in normal cases, as the application will most likely keep running/
	time.Sleep(time.Second * 30)
}
