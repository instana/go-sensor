package instagin

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	instana "github.com/instana/go-sensor"
)

func AddMiddleware(sensor *instana.Sensor, engine *gin.Engine) {
	f := middleware(sensor)
	engine.Handlers = append([]gin.HandlerFunc{f}, findAndRemove(f, engine.Handlers)...)

	// to trigger engine.rebuild404Handlers and engine.rebuild405Handlers
	engine.Use()
}

type statusWriter interface {
	SetStatus(status int)
}

var middleware = func(sensor *instana.Sensor) gin.HandlerFunc {
	return func(gc *gin.Context) {
		instana.TracingHandlerFunc(sensor, "", func(writer http.ResponseWriter, request *http.Request) {
			gc.Request = request
			gc.Next()

			if v, ok := writer.(statusWriter); ok {
				v.SetStatus(gc.Writer.Status())
			}
		})(gc.Writer, gc.Request)
	}
}

func findAndRemove(handler gin.HandlerFunc, handlers []gin.HandlerFunc) []gin.HandlerFunc {
	for k := range handlers {
		if reflect.ValueOf(handler).Pointer() == reflect.ValueOf(handlers[k]).Pointer() {
			return append(handlers[:k], handlers[k+1:]...)
		}
	}

	return handlers
}
