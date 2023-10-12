package common

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func SetCookie(c echo.Context, name, value, path string, expiry time.Time) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
		Expires:  expiry,
	})
}
