// SPDX-FileCopyrightText: 2026 IBM Corp.
// SPDX-FileCopyrightText: 2026 Instana Inc.
//
// SPDX-License-Identifier: MIT

package instaechov2

import (
	"net/http"

	instana "github.com/instana/go-sensor"
	"github.com/labstack/echo/v5"
)

// New returns an instrumented Echo v5 instance with Instana tracing middleware.
// The returned Echo instance will automatically create entry spans for all incoming HTTP requests.
func New(col instana.TracerLogger) *echo.Echo {
	e := echo.New()
	e.Use(Middleware(col))

	return e
}

// Middleware returns an Echo v5 middleware that instruments HTTP handlers with Instana tracing.
// It creates entry spans for incoming requests and propagates trace context to downstream handlers.
// The middleware should be added before defining routes to ensure all handlers are instrumented.
func Middleware(col instana.TracerLogger) echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) error {
			// Get route info directly from context
			ri := ctx.RouteInfo()

			// For unmatched routes (404), use request info as fallback
			if ri.Method == "" && ri.Path == "" {
				ri = echo.RouteInfo{
					Method: ctx.Request().Method,
					Path:   ctx.Path(),
				}
			}

			var err error

			instana.TracingNamedHandlerFunc(
				col,
				ri.Name,
				ri.Path,
				func(w http.ResponseWriter, req *http.Request) {
					ctx.SetResponse(w)
					ctx.SetRequest(req)

					if err = next(ctx); err != nil {
						ctx.Echo().HTTPErrorHandler(ctx, err)
					}
				})(ctx.Response(), ctx.Request())

			return err
		}
	})
}
