package main

import (
	"net/http"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

func RunnerHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*httperror.Error); ok {
		code = he.Code
	}
	sentry.CaptureException(err)
	c.JSON(code, echo.HTTPError{Message: err.Error()})
}
