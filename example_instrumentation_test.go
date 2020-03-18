package instana_test

import (
	"net/http"

	instana "github.com/instana/go-sensor"
)

// This example demonstrates how to instrument an HTTP handler with Instana and register it
// in http.DefaultServeMux
func ExampleTracingHandlerFunc() {
	// Here we initialize a new instance of instana.Sensor, however it is STRONGLY recommended
	// to use a single instance throughout your application
	sensor := instana.NewSensor("my-http-server")

	http.HandleFunc("/", instana.TracingHandlerFunc(sensor, "root", func(w http.ResponseWriter, req *http.Request) {
		// handler code
	}))
}
