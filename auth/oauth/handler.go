package oauth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/deepsourcecorp/runner/auth"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/ksuid"
	"golang.org/x/exp/slog"
)

type Handler struct {
	runner     *auth.Runner
	deepsource *auth.DeepSource
	factory    *Factory
	store      SessionStore
}

func NewHandler(runner *auth.Runner, deepsource *auth.DeepSource, store SessionStore, factory *Factory) *Handler {
	return &Handler{
		runner:     runner,
		deepsource: deepsource,
		factory:    factory,
		store:      store,
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

	token, err := backend.GetToken(ctx, req.Code)
	if err != nil {
		return c.JSON(500, err.Error())
	}

	session := NewSession()
	session.BackendToken = token
	if err := h.store.Create(session); err != nil {
		return c.JSON(500, err.Error())
	}

	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    session.ID,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})

	return c.Redirect(http.StatusTemporaryRedirect, "/apps/"+req.AppID+"/auth/session?state="+req.State)
}

type SessionRequest struct {
	AppID     string `param:"app_id"`
	State     string `query:"state"`
	SessionID string
}

func (r *SessionRequest) Parse(c echo.Context) error {
	req := new(SessionRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	r.AppID = req.AppID
	r.State = req.State

	cookie, err := c.Cookie("session")
	if err != nil {
		return err
	}
	if cookie == nil {
		return fmt.Errorf("session cookie not found")
	}
	r.SessionID = cookie.Value
	return nil
}

func (h *Handler) HandleSession(c echo.Context) error {
	req := new(SessionRequest)
	if err := req.Parse(c); err != nil {
		return c.JSON(400, err.Error())
	}

	session, err := h.store.GetByID(req.SessionID)
	if err != nil {
		if errors.Is(err, ErrNoSession) {
			return c.JSON(400, err.Error())
		}
		return c.JSON(500, err.Error())
	}
	code := ksuid.New().String()
	session.SetAccessCode(code)
	if err := h.store.Update(session); err != nil {
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

	session, err := h.store.GetByAccessCode(req.Code)
	if err != nil {
		slog.Error("error getting session", "error", err)
		return c.JSON(500, err.Error())
	}
	session.GenerateRunnerToken(session.BackendToken.Expiry)
	session.UnsetAccessCode()
	err = h.store.Update(session)
	if err != nil {
		slog.Error("error updating session", "error", err)
		return c.JSON(500, err.Error())
	}

	return c.JSON(200, session.RunnerToken)
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
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

func (h *Handler) HandleUser(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(UserRequest)
	if err := req.Parse(c); err != nil {
		return c.JSON(400, err.Error())
	}
	backend, err := h.factory.GetBackend(req.AppID)
	if err != nil {
		return c.JSON(400, err.Error())
	}

	session, err := h.store.GetByAccessToken(req.AccessToken)
	if err != nil {
		if errors.Is(err, ErrNoSession) {
			return c.JSON(400, err.Error())
		}
		return c.JSON(500, err.Error())
	}

	user, err := backend.GetUser(ctx, session.BackendToken)
	if err != nil {
		return c.JSON(500, err.Error())
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
	ctx := c.Request().Context()
	req := &RefreshRequest{}
	if err := c.Bind(req); err != nil {
		return c.JSON(400, err.Error())
	}

	if !(h.runner.IsValidClientID(req.ClientID) || !h.runner.IsValidClientSecret(req.ClientSecret)) {
		return c.JSON(400, "invalid client_id or client_secret")
	}

	backend, err := h.factory.GetBackend(req.AppID)
	if err != nil {
		return c.JSON(400, err.Error())
	}

	session, err := h.store.GetByRefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrNoSession) {
			return c.JSON(400, err.Error())
		}
		return c.JSON(500, err.Error())
	}

	backendToken, err := backend.RefreshToken(ctx, session.BackendToken.RefreshToken)
	if err != nil {
		return c.JSON(500, err.Error())
	}

	session.BackendToken = backendToken
	session.GenerateRunnerToken(session.BackendToken.Expiry)

	err = h.store.Update(session)
	if err != nil {
		return c.JSON(500, err.Error())
	}

	return nil
}
