// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.11

package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	instana "github.com/instana/go-sensor"

	"github.com/gorilla/mux"
	"github.com/instana/go-sensor/instrumentation/instagorillamux"
)

var listenAddr string

func main() {
	flag.StringVar(&listenAddr, "l", os.Getenv("LISTEN_ADDR"), "Server listen address")
	flag.Parse()

	if listenAddr == "" {
		flag.Usage()
		os.Exit(2)
	}

	r := mux.NewRouter()
	r.HandleFunc("/foo", func(writer http.ResponseWriter, request *http.Request) {})

	// create a sensor
	sensor := instana.NewSensor("gorillamux-sensor")

	// add middleware
	instagorillamux.AddMiddleware(sensor, r)

	log.Fatal(http.ListenAndServe(listenAddr, r))
}
