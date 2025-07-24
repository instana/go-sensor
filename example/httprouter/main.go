// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instahttprouter"
	"github.com/julienschmidt/httprouter"
)

var listenAddr string

func main() {
	flag.StringVar(&listenAddr, "l", os.Getenv("LISTEN_ADDR"), "Server listen address")
	flag.Parse()

	if listenAddr == "" {
		flag.Usage()
		os.Exit(2)
	}

	// Create a instana collector
	collector := instana.InitCollector(&instana.Options{
		Service: "my-web-server",
		Tracer:  instana.DefaultTracerOptions(),
	})

	// Create router and wrap it with Instana
	r := instahttprouter.Wrap(httprouter.New(), collector)

	// Define handlers
	r.GET("/foo", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {})
	r.Handle(http.MethodPost, "/foo/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {})

	// There is no need to additionally instrument your handlers with instana.TracingHandlerFunc(), since
	// the instrumented router takes care of this during the registration process.
	r.HandlerFunc(http.MethodDelete, "/foo/:id", func(writer http.ResponseWriter, request *http.Request) {})

	log.Fatal(http.ListenAndServe(listenAddr, r))
}
