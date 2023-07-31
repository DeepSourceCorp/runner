package token

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SessionAuthMiddleware(runnerID string, service *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			referer := c.Request().URL.String()
			c.Response().Header().Set("referer", referer)

			cookie, err := c.Cookie("session")
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/refresh?redirect="+referer)
			}
			if cookie.Value == "" {
				return c.Redirect(http.StatusTemporaryRedirect, "/refresh?redirect="+referer)
			}
			_, err = service.ReadToken(runnerID, ScopeCodeRead, cookie.Value)
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/refresh?redirect="+referer)
			}
			return next(c)
		}
	}
}
