package middleware

import (
	"strings"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/jwtutil"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

func DeepSourceMiddleware(runnerID string, verifier *jwtutil.Verifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authrorization := c.Request().Header.Get("Authorization")
			parts := strings.SplitN(authrorization, " ", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				slog.Error("middleware/bearer: not a valid bearer token", slog.Any("header", authrorization))
				return httperror.ErrUnauthorized(nil)
			}

			claims, err := verifier.Verify(parts[1])
			if err != nil {
				slog.Error("middleware/bearer: failed to verify token", slog.Any("err", err))
				return httperror.ErrUnauthorized(nil)
			}

			if claims["runner_id"] != runnerID {
				slog.Error("middleware/bearer: runner id mismatch", slog.Any("runner-id", claims["runner-id"]))
				return httperror.ErrUnauthorized(nil)
			}
			return next(c)
		}
	}
}
