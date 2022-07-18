// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instahttprouter

import (
	"context"
	"net/http"

	instana "github.com/instana/go-sensor"
	"github.com/julienschmidt/httprouter"
)

// Wrap returns an instrumented instance of a httprouter.Router that
// instruments HTTP handlers with Instana upon registration.
func Wrap(r *httprouter.Router, sensor *instana.Sensor) *WrappedRouter {
	return &WrappedRouter{
		Router: r,
		sensor: sensor,
	}
}

type WrappedRouter struct {
	*httprouter.Router
	sensor *instana.Sensor
}

// GET is a shortcut for router.Handle(http.MethodGet, path, handle)
func (r *WrappedRouter) GET(path string, handle httprouter.Handle) {
	r.Handle(http.MethodGet, path, handle)
}

// HEAD is a shortcut for router.Handle(http.MethodHead, path, handle)
func (r *WrappedRouter) HEAD(path string, handle httprouter.Handle) {
	r.Handle(http.MethodHead, path, handle)
}

// OPTIONS is a shortcut for router.Handle(http.MethodOptions, path, handle)
func (r *WrappedRouter) OPTIONS(path string, handle httprouter.Handle) {
	r.Handle(http.MethodOptions, path, handle)
}

// POST is a shortcut for router.Handle(http.MethodPost, path, handle)
func (r *WrappedRouter) POST(path string, handle httprouter.Handle) {
	r.Handle(http.MethodPost, path, handle)
}

// PUT is a shortcut for router.Handle(http.MethodPut, path, handle)
func (r *WrappedRouter) PUT(path string, handle httprouter.Handle) {
	r.Handle(http.MethodPut, path, handle)
}

// PATCH is a shortcut for router.Handle(http.MethodPatch, path, handle)
func (r *WrappedRouter) PATCH(path string, handle httprouter.Handle) {
	r.Handle(http.MethodPatch, path, handle)
}

// DELETE is a shortcut for router.Handle(http.MethodDelete, path, handle)
func (r *WrappedRouter) DELETE(path string, handle httprouter.Handle) {
	r.Handle(http.MethodDelete, path, handle)
}

// Handle instruments and registers a new request handle with the given path and method.
//
// For details please refer to the (*httprouter.Router).Handle() documentation:
// https://pkg.go.dev/github.com/julienschmidt/httprouter#Router.Handle
func (r *WrappedRouter) Handle(method, path string, handle httprouter.Handle) {
	r.Router.HandlerFunc(method, path, instana.TracingHandlerFunc(r.sensor, path, func(w http.ResponseWriter, req *http.Request) {
		handle(w, req, httprouter.ParamsFromContext(req.Context()))
	}))
}

// Handler is an adapter which allows the usage of an uninstrumented http.Handler as a
// request handle.
// The Params are available in the request context under ParamsKey.
func (r *WrappedRouter) Handler(method, path string, handler http.Handler) {
	r.Handle(method, path,
		func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
			if len(p) > 0 {
				ctx := req.Context()
				ctx = context.WithValue(ctx, httprouter.ParamsKey, p)
				req = req.WithContext(ctx)
			}
			handler.ServeHTTP(w, req)
		},
	)
}

// HandlerFunc is an adapter which allows the usage of an uninstrumented http.HandlerFunc as a
// request handle.
func (r *WrappedRouter) HandlerFunc(method, path string, handler http.HandlerFunc) {
	r.Handler(method, path, handler)
}
