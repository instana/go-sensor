package main


import (
"fmt"
"net/http"

instana "github.com/instana/go-sensor"
)

//simple handler used for the example server

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from the OpenTelemetry POC")
}

func main() {
	//wrap the handler with the OpenTelemetry middleware POC
	handler := instana.OTelTracingHandlerFunc(
		"/",
		helloHandler,
	)

	//Register the endpoint
	http.HandleFunc("/", handler)

	fmt.Println("Listening on :8081")

	//Start with the HTTP server
	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}
}