package main

import (
	"net/http"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

func RunnerHTTPErrorHandler(err error, c echo.Context) {
	sentry.CaptureException(err)
	switch typedErr := err.(type) {
	case *httperror.Error:
		_ = c.JSON(typedErr.Code, typedErr)
	default:
		_ = c.JSON(http.StatusInternalServerError, httperror.ErrUnknown(err))
	}
}
