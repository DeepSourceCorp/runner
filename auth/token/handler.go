package token

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) HandleRefresh(c echo.Context) error {
	referrer := c.Request().Referer()
	cookie, err := c.Cookie("refresh")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}
	if cookie.Value == "" {
		return c.JSON(http.StatusUnauthorized, "invalid refresh token")
	}

	user, err := h.service.ReadRefreshToken(cookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}

	accessToken, err := h.service.GetAccessToken(user)
	if err != nil {
		return c.JSON(500, err.Error())
	}
	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    accessToken,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})

	return c.Redirect(http.StatusTemporaryRedirect, referrer)
}
