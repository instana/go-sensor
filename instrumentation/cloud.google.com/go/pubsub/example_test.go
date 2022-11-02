// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package pubsub_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub"
)

// This example show how to instrument and HTTP handler that receives Google Cloud Pub/Sub messages
// via the push delivery method.
func ExampleTracingHandlerFunc() {
	// Initialize sensor
	sensor := instana.NewSensor("pubsub-consumer")

	// Wrap your Pub/Sub message handler with pubsub.TracingHandlerFunc
	http.Handle("/", pubsub.TracingHandlerFunc(sensor, "/", func(w http.ResponseWriter, req *http.Request) {
		var delivery struct {
			Message struct {
				Data []byte `json:"data"`
			} `son:"message"`
		}

		if err := json.NewDecoder(req.Body).Decode(&delivery); err != nil {
			// handle the error
		}

		fmt.Printf("got %q", delivery.Message.Data)
	}))
}
