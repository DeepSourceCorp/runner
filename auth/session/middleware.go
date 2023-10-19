package session

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Middleware struct {
	service *Service
}

func NewMiddleware(sessionService *Service) *Middleware {
	return &Middleware{
		service: sessionService,
	}
}

func (m *Middleware) HandleSesionAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		referrer := c.Request().URL.String()
		c.Response().Header().Set("referrer", referrer)

		cookie, err := c.Cookie("session")
		if err != nil {
			return err
		}
		if cookie.Value == "" {
			return echo.ErrUnauthorized
		}

		session, err := m.service.FetchSessionByJWT(cookie.Value, ScopeCode)
		if err != nil {
			return err
		}

		if err != nil || session == nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/refresh?redirect="+referrer)
		}

		return next(c)
	}
}
