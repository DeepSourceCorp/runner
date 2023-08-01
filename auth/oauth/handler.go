package oauth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/deepsourcecorp/runner/auth/store"
	"github.com/deepsourcecorp/runner/auth/token"
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
	req := new(AuthorizationRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(400, err.Error())
	}
	scopes := strings.Split(req.Scopes, " ")

	if !h.runner.IsValidClientID(req.ClientID) {
		return c.JSON(400, fmt.Sprintf("invalid client_id: %s", req.ClientID))
	}

	backend, err := h.factory.GetBackend(req.AppID)
	if err != nil {
		return c.JSON(400, err.Error())
	}

	url, err := backend.AuthorizationURL(req.State, scopes)
	if err != nil {
		return c.JSON(500, err.Error())
	}
	return c.Redirect(http.StatusTemporaryRedirect, url)
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
		return c.JSON(400, err.Error())
	}

	backend, err := h.factory.GetBackend(req.AppID)
	if err != nil {
		return c.JSON(400, err.Error())
	}

	t, err := backend.GetToken(ctx, req.Code)
	if err != nil {
		return c.JSON(500, err.Error())
	}

	user, err := backend.GetUser(ctx, t)
	if err != nil {
		return c.JSON(500, err.Error())
	}

	accessToken, err := h.tokenService.GenerateToken(h.runner.ID, []string{token.ScopeUser, token.ScopeCodeRead}, user, token.ExpiryAccessToken)
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

	refreshToken, err := h.tokenService.GenerateToken(h.runner.ID, []string{token.ScopeRefresh}, user, token.ExpiryRefreshToken)
	if err != nil {
		return c.JSON(500, err.Error())
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh",
		Value:    refreshToken,
		Path:     "/refresh",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})

	return c.Redirect(http.StatusTemporaryRedirect, "/apps/"+req.AppID+"/auth/session?state="+req.State)
}

type SessionRequest struct {
	AppID string `param:"app_id"`
	State string `query:"state"`
}

func (h *Handler) HandleSession(c echo.Context) error {
	req := SessionRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, err.Error())
	}

	cookie, err := c.Cookie("session")
	if err != nil {
		return c.JSON(400, err.Error())
	}

	user, err := h.tokenService.ReadToken(h.runner.ID, token.ScopeUser, cookie.Value)
	if err != nil {
		return c.JSON(400, err.Error())
	}

	code := ksuid.New().String()
	if err := h.store.SetAccessCode(code, user); err != nil {
		return c.JSON(500, err.Error())
	}

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
		slog.Error("error binding request", "error", err)
		return c.JSON(400, err.Error())
	}
	if !h.runner.IsValidClientID(req.ClientID) || !h.runner.IsValidClientSecret(req.ClientSecret) {
		return c.JSON(400, "invalid client_id or client_secret")
	}

	user, err := h.store.VerifyAccessCode(req.Code)
	if err != nil {
		return c.JSON(http.StatusForbidden, err.Error())
	}

	accessToken, err := h.tokenService.GenerateToken(h.runner.ID, []string{token.ScopeUser}, user, token.ExpiryAccessToken)
	if err != nil {
		return c.JSON(500, err.Error())
	}
	refreshtToken, err := h.tokenService.GenerateToken(h.runner.ID, []string{token.ScopeUser}, user, token.ExpiryRefreshToken)
	if err != nil {
		return c.JSON(500, err.Error())
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
		return c.JSON(400, err.Error())
	}

	user, err := h.tokenService.ReadToken(h.runner.ID, token.ScopeUser, req.AccessToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
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
		return c.JSON(400, err.Error())
	}

	if !(h.runner.IsValidClientID(req.ClientID) || !h.runner.IsValidClientSecret(req.ClientSecret)) {
		return c.JSON(400, "invalid client_id or client_secret")
	}

	user, err := h.tokenService.ReadToken(h.runner.ID, token.ScopeRefresh, req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}

	accessToken, err := h.tokenService.GenerateToken(h.runner.ID, []string{token.ScopeUser}, user, token.ExpiryAccessToken)
	if err != nil {
		return c.JSON(500, err.Error())
	}

	return c.JSON(200, &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
		Expiry:       time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	})
}
