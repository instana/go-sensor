// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

//go:build go1.12
// +build go1.12

package instamux

import (
	"net/http"

	"github.com/gorilla/mux"
	instana "github.com/instana/go-sensor"
)

// NewRouter is wrapper for mux.NewRouter()
func NewRouter(sensor *instana.Sensor) *mux.Router {
	r := mux.NewRouter()
	AddMiddleware(sensor, r)

	return r
}

// AddMiddleware instruments the mux.Router instance with Instana
func AddMiddleware(sensor *instana.Sensor, router *mux.Router) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			r := mux.CurrentRoute(req)

			pathTemplate, err := r.GetPathTemplate()
			if err != nil {
				sensor.Logger().Debug("can not get path template from the route: ", err)
				pathTemplate = ""
			}

			instana.TracingNamedHandlerFunc(sensor, r.GetName(), pathTemplate, func(w http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(w, req)
			})(w, req)
		})
	})
}
