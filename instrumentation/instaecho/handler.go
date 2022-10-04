// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

//go:build go1.16
// +build go1.16

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
			r := lookupMatchedRoute(c)
			if r == nil {
				r = &echo.Route{
					Method: c.Request().Method,
					Path:   c.Path(),
				}
			}

			var err error

			instana.TracingNamedHandlerFunc(sensor, r.Name, r.Path, func(w http.ResponseWriter, req *http.Request) {
				c.SetResponse(echo.NewResponse(w, c.Echo()))
				c.SetRequest(req)

				if err = next(c); err != nil {
					c.Error(err)
				}

			})(c.Response(), c.Request())

			return err
		}
	}
}

func lookupMatchedRoute(c echo.Context) *echo.Route {
	path := c.Path()

	for _, r := range c.Echo().Routes() {
		if r.Path == path {
			return r
		}
	}

	return nil
}
