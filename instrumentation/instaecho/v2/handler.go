// (c) Copyright IBM Corp. 2026
// (c) Copyright Instana Inc. 2026

package instaechov2

import "github.com/labstack/echo/v5"

func New() *echo.Echo {
	e := echo.New()
	e.Use(Middleware())

	return e
}

func Middleware() echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Tracing logic
			return next(c)
		}
	})
}
