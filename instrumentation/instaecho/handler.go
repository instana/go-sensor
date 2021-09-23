// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.13

package instaecho

import (
	"net/http"

	instana "github.com/instana/go-sensor"
	"github.com/labstack/echo/v4"
)

// AddMiddleware adds the tracing middleware to the list of Echo handlers.
func AddMiddleware(sensor *instana.Sensor, engine *echo.Echo) {
	f := middleware(sensor)
	engine.Use(f)
}

// middleware wraps Echo's handlers execution. Adds tracing context and handles entry span.
func middleware(sensor *instana.Sensor) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var err error

			instana.TracingHandlerFunc(sensor, c.Path(), func(writer http.ResponseWriter, request *http.Request) {
				c.SetResponse(echo.NewResponse(writer, c.Echo()))
				c.SetRequest(request)

				if err = next(c); err != nil {
					c.Error(err)
				}

			})(c.Response(), c.Request())

			return err
		}
	}
}
