package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

type SessionStore interface {
	IsValidSession(id string) bool
}

type SessionAuthenticationMiddleware struct {
	oauthSessionStore SessionStore
}

func NewSessionMiddleware(oauthSessionStore SessionStore) *SessionAuthenticationMiddleware {
	return &SessionAuthenticationMiddleware{
		oauthSessionStore: oauthSessionStore,
	}
}

func (m *SessionAuthenticationMiddleware) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionID, err := extractSessionID(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err.Error())
		}
		if m.oauthSessionStore != nil && m.oauthSessionStore.IsValidSession(sessionID) {
			return next(c)
		}
		return c.JSON(http.StatusUnauthorized, "invalid session")
	}
}

func extractSessionID(c echo.Context) (string, error) {
	cookie, err := c.Cookie("session")
	if err != nil {
		slog.Warn("error extracting session cookie", slog.Any("err", err))
		return "", errors.New("invalid session")
	}

	if cookie == nil || cookie.Value == "" {
		slog.Warn("session cookie not found", slog.Any("err", err))
		return "", errors.New("invalid session")
	}
	return cookie.Value, nil
}
