package token

import (
	"net/http"

	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
	runner  *model.Runner
}

func NewHandler(runner *model.Runner, service *Service) *Handler {
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

	user, err := h.service.ReadToken(h.runner.ID, ScopeRefresh, cookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}

	accessToken, err := h.service.GenerateToken(h.runner.ID, []string{ScopeUser, ScopeCodeRead}, user)
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

func (h *Handler) HandleLogout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})

	c.SetCookie(&http.Cookie{
		Name:     "refresh",
		Value:    "",
		Path:     "/refresh",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})

	return c.NoContent(http.StatusOK)
}
