package oauth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/deepsourcecorp/runner/auth/store"
	"github.com/deepsourcecorp/runner/auth/token"
	"github.com/deepsourcecorp/runner/httperror"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/ksuid"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

type Handler struct {
	runner     *model.Runner
	deepsource *model.DeepSource
	factory    *Factory
	store      store.Store

	tokenService *token.Service
}

var ErrInvalidClientCredentials = errors.New("invalid client credentials")

func NewHandler(runner *model.Runner, deepsource *model.DeepSource, store store.Store, tokenService *token.Service, factory *Factory) *Handler {
	return &Handler{
		runner:     runner,
		deepsource: deepsource,
		factory:    factory,
		store:      store,

		tokenService: tokenService,
	}
}

type AuthorizationRequest struct {
	AppID    string `param:"app_id"`
	ClientID string `query:"client_id"`
	Scopes   string `query:"scopes"`
	State    string `query:"state"`
}

func (h *Handler) HandleAuthorize(c echo.Context) error {
	// Parse the authorization request.
	req := new(AuthorizationRequest)
	if err := c.Bind(req); err != nil {
		slog.Error("authorization request bind error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}

	scopes := strings.Split(req.Scopes, " ")

	if !h.runner.IsValidClientID(req.ClientID) {
		slog.Warn("authorization request with invalid client id", slog.Any("client_id", req.ClientID))
		return httperror.ErrAppInvalid(errors.New("invalid client id"))
	}

	// Generate the authroization URL for the upstream identity provider and
	// redirect the user to the authorization URL for the identity provider.
	backend, err := h.factory.GetBackend(req.AppID)
	if err != nil {
		slog.Warn("authorization request on app with unsupported provider", slog.Any("app_id", req.AppID))
		return httperror.ErrAppUnsupported(err)
	}

	redirectURL := backend.AuthorizationURL(req.State, scopes)
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

type CallbackRequest struct {
	AppID string `param:"app_id"`
	Code  string `query:"code"`
	State string `query:"state"`
}

func (h *Handler) HandleCallback(c echo.Context) error {
	ctx := c.Request().Context()

	req := &CallbackRequest{}
	if err := c.Bind(req); err != nil {
		slog.Error("callback request bind error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}

	backend, err := h.factory.GetBackend(req.AppID)
	if err != nil {
		slog.Warn("callback request on app with unsupported provider", slog.Any("app_id", req.AppID))
		return httperror.ErrAppUnsupported(err)
	}

	t, err := backend.GetToken(ctx, req.Code)
	if err != nil {
		slog.Error("callback request failed while getting token", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}

	user, err := backend.GetUser(ctx, t)
	if err != nil {
		slog.Error("callback request failed while getting user", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}

	// Generate and set the access token cookie.  This will be used during the
	// session request to authenticate the authentication request.
	accessToken, err := h.tokenService.GenerateToken(
		h.runner.ID, []string{token.ScopeUser, token.ScopeCodeRead},
		user, token.ExpiryAccessToken,
	)
	if err != nil {
		slog.Error("callback request failed while generating access token", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    accessToken,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})

	// Generate and set the refresh token cookie.
	refreshToken, err := h.tokenService.GenerateToken(
		h.runner.ID, []string{token.ScopeRefresh},
		user, token.ExpiryRefreshToken)
	if err != nil {
		slog.Error("callback request failed while generating refresh token", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh",
		Value:    refreshToken,
		Path:     "/refresh",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})

	// We redirect to the session endpoint to set the session cookie.
	// This breaks the OAuth2 convention, however this is a necessary
	// evil to set a session between the user and the Runner instance.
	return c.Redirect(http.StatusTemporaryRedirect, "/apps/"+req.AppID+"/auth/session?state="+req.State)
}

type SessionRequest struct {
	AppID string `param:"app_id"`
	State string `query:"state"`
}

func (h *Handler) HandleSession(c echo.Context) error {
	req := SessionRequest{}
	if err := c.Bind(&req); err != nil {
		slog.Error("session request bind error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}

	cookie, err := c.Cookie("session")
	if err != nil {
		slog.Error("session request failed while getting session cookie", slog.Any("err", err))
		return httperror.ErrUnauthorized(err)
	}

	user, err := h.tokenService.ReadToken(h.runner.ID, token.ScopeUser, cookie.Value)
	if err != nil {
		slog.Error("session request failed while reading session token", slog.Any("err", err))
		return httperror.ErrUnauthorized(err)
	}

	code := ksuid.New().String()
	if err := h.store.SetAccessCode(code, user); err != nil {
		slog.Error("session request failed while setting access code", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}

	// Redirect back to DeepSource as the authorization callback with the code.
	u := h.deepsource.Host.JoinPath(fmt.Sprintf("/accounts/runner/apps/%s/login/callback/bifrost/", req.AppID))
	q := u.Query()
	q.Add("app_id", req.AppID)
	q.Add("code", code)
	q.Add("state", req.State)
	u.RawQuery = q.Encode()

	return c.Redirect(http.StatusTemporaryRedirect, u.String())
}

type TokenRequest struct {
	AppID        string `param:"app_id"`
	Code         string `query:"code" json:"code"`
	ClientID     string `query:"client_id" json:"client_id"`
	ClientSecret string `query:"client_secret" json:"client_secret"`
}

func (h *Handler) HandleToken(c echo.Context) error {
	req := new(TokenRequest)
	if err := c.Bind(req); err != nil {
		slog.Error("token request bind error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}
	if !h.runner.IsValidClientID(req.ClientID) || !h.runner.IsValidClientSecret(req.ClientSecret) {
		slog.Warn("token request with invalid client credentials", slog.Any("client_id", req.ClientID))
		return httperror.ErrAppInvalid(ErrInvalidClientCredentials)
	}

	user, err := h.store.VerifyAccessCode(req.Code)
	if err != nil {
		slog.Error("token request failed while verifying access code", slog.Any("err", err))
		return httperror.ErrUnauthorized(err)
	}

	accessToken, err := h.tokenService.GenerateToken(h.runner.ID, []string{token.ScopeUser}, user, token.ExpiryAccessToken)
	if err != nil {
		slog.Error("token request failed while generating access token", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	refreshtToken, err := h.tokenService.GenerateToken(h.runner.ID, []string{token.ScopeUser}, user, token.ExpiryRefreshToken)
	if err != nil {
		slog.Error("token request failed while generating refresh token", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}

	return c.JSON(200, &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshtToken,
		Expiry:       time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	})
}

type UserRequest struct {
	AppID       string
	AccessToken string
}

func (r *UserRequest) Parse(c echo.Context) error {
	r.AppID = c.Param("app_id")
	h := c.Request().Header.Get("Authorization")
	if h == "" {
		return fmt.Errorf("missing authorization header")
	}
	parts := strings.Split(h, " ")
	if len(parts) != 2 {
		return fmt.Errorf("invalid authorization header")
	}
	if parts[0] != "Bearer" {
		return fmt.Errorf("invalid authorization header")
	}
	r.AccessToken = parts[1]
	return nil
}

type UserResponse struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

func (h *Handler) HandleUser(c echo.Context) error {
	req := new(UserRequest)
	if err := req.Parse(c); err != nil {
		slog.Error("user request parse error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}

	user, err := h.tokenService.ReadToken(h.runner.ID, token.ScopeUser, req.AccessToken)
	if err != nil {
		slog.Error("user request failed while reading access token", slog.Any("err", err))
		return httperror.ErrUnauthorized(err)
	}

	return c.JSON(200, &UserResponse{
		UserName: user.Login,
		Email:    user.Email,
		FullName: user.Name,
		ID:       user.ID,
	})
}

type RefreshRequest struct {
	AppID        string `param:"app_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) HandleRefresh(c echo.Context) error {
	req := &RefreshRequest{}
	if err := c.Bind(req); err != nil {
		slog.Error("refresh request bind error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}

	if !(h.runner.IsValidClientID(req.ClientID) || !h.runner.IsValidClientSecret(req.ClientSecret)) {
		slog.Warn("refresh request with invalid client credentials", slog.Any("client_id", req.ClientID))
		return httperror.ErrAppInvalid(ErrInvalidClientCredentials)
	}

	user, err := h.tokenService.ReadToken(h.runner.ID, token.ScopeRefresh, req.RefreshToken)
	if err != nil {
		slog.Error("refresh request failed while reading refresh token", slog.Any("err", err))
		return httperror.ErrUnauthorized(err)
	}

	accessToken, err := h.tokenService.GenerateToken(h.runner.ID, []string{token.ScopeUser}, user, token.ExpiryAccessToken)
	if err != nil {
		slog.Error("refresh request failed while generating access token", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}

	return c.JSON(200, &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
		Expiry:       time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	})
}
