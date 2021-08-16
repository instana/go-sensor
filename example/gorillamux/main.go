// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.11

package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instamux"
)

var listenAddr string

func main() {
	flag.StringVar(&listenAddr, "l", os.Getenv("LISTEN_ADDR"), "Server listen address")
	flag.Parse()

	if listenAddr == "" {
		flag.Usage()
		os.Exit(2)
	}

	// create a sensor
	sensor := instana.NewSensor("my-web-server")

	// create router
	r := mux.NewRouter()

	// instrument your router by adding a middleware
	instamux.AddMiddleware(sensor, r)

	r.HandleFunc("/foo", func(writer http.ResponseWriter, request *http.Request) {})

	log.Fatal(http.ListenAndServe(listenAddr, r))
}
