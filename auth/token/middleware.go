package token

import (
	"net/http"
	"strings"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/labstack/echo/v4"
)

var ErrInvalidToken = httperror.Error{Message: "invalid token"}

func BearerAuthMiddleware(service *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authrorization := c.Request().Header.Get("Authorization")
			parts := strings.SplitN(authrorization, " ", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return c.JSON(http.StatusUnauthorized, ErrInvalidToken)
			}
			_, err := service.ReadAccessToken(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, ErrInvalidToken)
			}
			// Token is valid
			return next(c)
		}
	}
}

func SessionAuthMiddleware(service *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("session")
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/refresh")
			}
			if cookie.Value == "" {
				return c.Redirect(http.StatusTemporaryRedirect, "/refresh")
			}
			_, err = service.ReadAccessToken(cookie.Value)
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/refresh")
			}
			return next(c)
		}
	}
}
