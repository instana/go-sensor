package instagin

import (
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
		htspan := instana.NewHttpSpan(gc.Request, sensor, "")

		defer htspan.Finish()
		defer func() {
			// Be sure to capture any kind of panic/error
			if err := recover(); err != nil {
				htspan.CollectPanicInformation(err)

				// re-throw the panic
				panic(err)
			}
		}()

		htspan.Inject(gc.Writer)

		gc.Next()

		htspan.CollectResponseHeaders(gc.Writer)
		htspan.CollectResponseStatus(gc.Writer)
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
