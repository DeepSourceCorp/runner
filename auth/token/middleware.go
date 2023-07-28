package token

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SessionAuthMiddleware(runnerID string, service *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("session")
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/refresh")
			}
			if cookie.Value == "" {
				return c.Redirect(http.StatusTemporaryRedirect, "/refresh")
			}
			_, err = service.ReadAccessToken(runnerID, cookie.Value)
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/refresh")
			}
			return next(c)
		}
	}
}
