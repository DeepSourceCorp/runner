package session

import (
	"time"

	"github.com/deepsourcecorp/runner/auth/common"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func (h *Handler) HandleRefresh(c echo.Context) error {
	req, err := NewRefreshRequest(c)
	if err != nil {
		return err
	}

	session, err := h.service.FetchSessionByJWT(req.RefreshToken, ScopeRefresh)
	if err != nil {
		return err
	}

	session, err = h.service.RefreshOAuthToken(session)
	if err != nil {
		return err
	}

	expiryAccessToken := time.Now().Add(common.ExpiryRunnerAccessToken)
	common.SetCookie(c, common.CookieNameSession, session.RunnerAccessToken, "/", expiryAccessToken)

	expiryRefreshToken := time.Now().Add(common.ExpiryRunnerRefreshToken)
	common.SetCookie(c, common.CookieNameRefresh, session.RunnerRefreshToken, "/refresh", expiryRefreshToken)

	return nil
}
