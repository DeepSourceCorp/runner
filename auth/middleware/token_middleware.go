package middleware

import (
	"net/http"

	"github.com/deepsourcecorp/runner/auth"
	"github.com/labstack/echo/v4"
)

type TokenMiddleware struct {
	Verifier *auth.Verifier
	RunnerID string
}

func NewTokenMiddleware(runnerID string, verifier *auth.Verifier) *TokenMiddleware {
	return &TokenMiddleware{
		Verifier: verifier,
		RunnerID: runnerID,
	}
}

func (m *TokenMiddleware) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
		}

		// Verify the token
		claims, err := m.Verifier.Verify(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		runnerID, ok := claims["runner_id"].(string)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		if runnerID != m.RunnerID {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		// Token is valid
		return next(c)
	}
}
