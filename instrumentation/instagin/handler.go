package instagin

import (
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

// middleware wraps gin's handlers execution. Adds tracing context and handles entry span.
var middleware = func(sensor *instana.Sensor) gin.HandlerFunc {
	return func(gc *gin.Context) {
		httpSpan := instana.NewHttpEntrySpan(gc.Request, sensor, "")

		// ensure that Finish() is a last call
		defer httpSpan.Finish()
		defer func() {
			// Be sure to capture any kind of panic/error
			if err := recover(); err != nil {
				httpSpan.CollectPanicInformation(err)

				// re-throw the panic
				panic(err)
			}
		}()

		httpSpan.Inject(gc.Writer)
		gc.Request = httpSpan.RequestWithContext(gc.Request)

		gc.Next()

		httpSpan.CollectResponseHeaders(gc.Writer)
		httpSpan.CollectResponseStatus(gc.Writer.Status())
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
