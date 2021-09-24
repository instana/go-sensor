// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.15

package instaecho

import (
	"net/http"

	instana "github.com/instana/go-sensor"
	"github.com/labstack/echo/v4"
)

// New returns an instrumented Echo.
func New(sensor *instana.Sensor) *echo.Echo {
	engine := echo.New()
	engine.Use(Middleware(sensor))

	return engine
}

// Middleware wraps Echo's handlers execution. Adds tracing context and handles entry span.
// It should be added as a first Middleware to the Echo, before defining handlers.
func Middleware(sensor *instana.Sensor) echo.MiddlewareFunc {
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
