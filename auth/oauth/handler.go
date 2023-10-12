package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/deepsourcecorp/runner/auth/common"
	"github.com/deepsourcecorp/runner/auth/contract"
	"github.com/deepsourcecorp/runner/httperror"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/ksuid"
	"golang.org/x/exp/slog"
)

const (

	// SessionURLFmt is the callback path for the interstitial page for setting
	// the session cookie.
	SessionURLFmt = "/apps/%s/auth/session?state=%s"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service}
}

func (h *Handler) HandleAuthorize(c echo.Context) error {
	req, err := contract.NewAuthorizationRequest(c)
	if err != nil {
		return httperror.ErrBadRequest(err)
	}
	url, err := h.service.GetAuthorizationURL(req)
	if err != nil {
		return err
	}

	return c.Redirect(302, url)
}

func (h *Handler) HandleCallback(c echo.Context) error {
	req, err := NewCallbackRequest(c)
	if err != nil {
		return httperror.ErrBadRequest(err)
	}

	session, err := h.service.CreateSession(req)
	if err != nil {
		return err
	}

	expiryAccessToken := time.Now().Add(common.ExpiryRunnerAccessToken)
	common.SetCookie(c, common.CookieNameSession, session.RunnerAccessToken, "/", expiryAccessToken)

	expiryRefreshToken := time.Now().Add(common.ExpiryRunnerRefreshToken)
	common.SetCookie(c, common.CookieNameRefresh, session.RunnerRefreshToken, "/refresh", expiryRefreshToken)

	return c.Redirect(
		http.StatusTemporaryRedirect,
		fmt.Sprintf(SessionURLFmt, req.AppID, req.State),
	)
}

func (h *Handler) HandleSession(c echo.Context) error {
	req, err := NewSessionRequest(c)
	if err != nil {
		slog.Error("failed to parse session request", slog.Any("err", err))
		return httperror.ErrBadRequest(err)
	}
	session, err := h.service.GenerateAccessCode(req)
	if err != nil {
		slog.Error("failed to generate access code", slog.Any("err", errors.Unwrap(err)))
		return err
	}

	url := h.service.DeepSourceCallbackURL(req.AppID,
		url.Values{
			"code":   {session.Code},
			"state":  {req.State},
			"app_id": {req.AppID},
		},
	)

	slog.Debug("redirecting to deepsource callback url", slog.Any("url", url))

	return c.Redirect(
		http.StatusTemporaryRedirect, url)
}

func (h *Handler) HandleToken(c echo.Context) error {
	req, err := NewTokenRequest(c)
	if err != nil {
		return httperror.ErrBadRequest(err)
	}

	session, err := h.service.GenerateOAuthToken(req)
	if err != nil {
		return err
	}

	res := &TokenResponse{
		AccessToken:  session.RunnerAccessToken,
		TokenType:    "bearer",
		ExpiresIn:    time.Now().Unix() - session.RunnerTokenExpiry,
		Expiry:       time.Unix(session.RunnerTokenExpiry, 0),
		RefreshToken: session.RunnerRefreshToken,
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) HandleRefresh(c echo.Context) error {
	req, err := contract.NewRefreshRequest(c)
	if err != nil {
		return httperror.ErrBadRequest(err)
	}

	session, err := h.service.RefreshOAuthToken(req)
	if err != nil {
		return err
	}

	res := &TokenResponse{
		AccessToken:  session.RunnerAccessToken,
		TokenType:    "bearer",
		ExpiresIn:    time.Now().Unix() - session.RunnerTokenExpiry,
		Expiry:       time.Unix(session.RunnerTokenExpiry, 0),
		RefreshToken: session.RunnerRefreshToken,
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) HandleUser(c echo.Context) error {
	ctx := context.WithValue(c.Request().Context(), common.ContextKeyRequestID, ksuid.New().String())
	req, err := contract.NewUserRequest(c)
	if err != nil {
		common.Log(ctx, slog.LevelError, "failed to parse user request", slog.Any("err", err))
		return httperror.ErrBadRequest(err)
	}
	req.Ctx = ctx

	user, err := h.service.GetUser(req)
	if err != nil {
		common.Log(ctx, slog.LevelError, "failed to get user", slog.Any("err", err))
		return err
	}

	return c.JSON(http.StatusOK, &contract.UserResponse{
		UserName: user.Login,
		Email:    user.Email,
		FullName: user.Name,
		ID:       user.ID,
	})
}
