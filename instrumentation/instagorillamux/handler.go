// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.11

package instagorillamux

import (
	"net/http"

	"github.com/gorilla/mux"
	instana "github.com/instana/go-sensor"
)

func AddMiddleware(sensor *instana.Sensor, router *mux.Router) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pathTemplate, err := mux.CurrentRoute(r).GetPathTemplate()
			if err != nil {
				pathTemplate = ""
			}

			instana.TracingHandlerFunc(sensor, pathTemplate, func(writer http.ResponseWriter, request *http.Request) {
				next.ServeHTTP(writer, request)
			})(w, r)
		})
	})
}
