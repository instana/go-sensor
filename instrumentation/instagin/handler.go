// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.11

package instagin

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	instana "github.com/instana/go-sensor"
)

// AddMiddleware adds Middleware to the gin Handlers list. Unlike method Use from the current gin API, it adds
// to the beginning of the list. It will allows to trace all the default handlers added during gin.Default() call.
func AddMiddleware(sensor *instana.Sensor, engine *gin.Engine) {
	f := middleware(sensor)
	engine.Handlers = append([]gin.HandlerFunc{f}, tryFindAndRemove(f, engine.Handlers)...)

	// trigger engine.rebuild404Handlers and engine.rebuild405Handlers
	engine.Use()
}

type statusWriter interface {
	SetStatus(status int)
}

// middleware wraps gin's handlers execution. Adds tracing context and handles entry span.
var middleware = func(sensor *instana.Sensor) gin.HandlerFunc {
	return func(gc *gin.Context) {
		instana.TracingHandlerFunc(sensor, "", func(writer http.ResponseWriter, request *http.Request) {
			gc.Request = request
			gc.Next()

			// set status from gc.Writer to instana.statusCodeRecorder which is used by instana.TracingHandlerFunc
			if v, ok := writer.(statusWriter); ok {
				v.SetStatus(gc.Writer.Status())
			}
		})(gc.Writer, gc.Request)
	}
}

// tryFindAndRemove tries to find a previously registered middleware and remove it from the handlers list.
// This function not necessarily is able to find duplicates. See documentation for a Pointer() method.
func tryFindAndRemove(handler gin.HandlerFunc, handlers []gin.HandlerFunc) []gin.HandlerFunc {
	for k := range handlers {
		if reflect.ValueOf(handler).Pointer() == reflect.ValueOf(handlers[k]).Pointer() {
			return append(handlers[:k], handlers[k+1:]...)
		}
	}

	return handlers
}
