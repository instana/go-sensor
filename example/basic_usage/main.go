// (c) Copyright IBM Corp. 2023

package main

import (
	"log"
	"net/http"

	instana "github.com/instana/go-sensor"
)

func main() {
	col := instana.InitCollector(&instana.Options{
		Service:           "Basic Usage",
		EnableAutoProfile: true,
		Tracer:            instana.DefaultTracerOptions(),
	})

	http.HandleFunc("/endpoint", instana.TracingHandlerFunc(col, "/endpoint", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	log.Fatal(http.ListenAndServe(":7070", nil))
}
