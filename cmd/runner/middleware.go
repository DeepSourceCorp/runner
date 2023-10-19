package main

import (
	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/jwtutil"
	"github.com/deepsourcecorp/runner/middleware"
	"github.com/labstack/echo/v4"
)

func DeepSourceMiddleware(c *config.Config) echo.MiddlewareFunc {
	verifier := jwtutil.NewVerifier(c.DeepSource.PublicKey)
	return middleware.DeepSourceMiddleware(c.Runner.ID, verifier)
}
