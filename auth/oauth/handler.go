package oauth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/deepsourcecorp/runner/auth/session"
	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/model"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const (
	// BifrostCallbackURLFmt is the callback path for DeepSource Cloud.
	BifrostCallbackURLFmt = "/accounts/runner/apps/%s/login/callback/bifrost/"

	// SessionURLFmt is the callback path for the interstitial page for setting
	// the session cookie.
	SessionURLFmt = "/apps/%s/auth/session?state=%s"
)

// Handler handles the OAuth2 flow for the Runner.  This handler exposes echo handlers.
type Handler struct {
	cloudURL       *url.URL
	runner         *model.Runner
	backend        *BackendFacade
	sessionService *session.Service
}

// NewHandler returns a new Handler instance.
func NewHandler(
	runner *model.Runner, cloudURL *url.URL, apps map[string]*App, sessionService *session.Service,
) (*Handler, error) {
	return &Handler{
		runner:         runner,
		cloudURL:       cloudURL,
		sessionService: sessionService,
		backend:        NewBackendFacade(apps),
	}, nil
}

// HandleAuthorize generates the authorization URL for the underlying backend
// and redirects the user to the authorization URL.  This triggers the login
// page on the OAuth backend.
func (h *Handler) HandleAuthorize(c echo.Context) error {
	req, err := NewAuthorizationRequest(c)
	if err != nil {
		return err
	}
	if !h.runner.IsValidClientID(req.ClientID) {
		return httperror.ErrUnknown(fmt.Errorf("invalid client id")) // TODO: Add ErrNotFound
	}
	authorizationURL, err := h.backend.AuthorizationURL(
		req.AppID, req.State, strings.Split(req.Scopes, ","),
	)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusTemporaryRedirect, authorizationURL)
}

// HandleCallback handles the callback request from the underlying backend.  This
// completes the OAuth flow with the backend and sets a session cookie that will
// be used to generate the runner token.
func (h *Handler) HandleCallback(c echo.Context) error {
	req, err := NewCallbackRequest(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	backendToken, err := h.backend.Exchange(ctx, req.AppID, req.State, req.Code)
	if err != nil {
		return err
	}
	session := session.New()
	if err := h.sessionService.SetBackendToken(
		session, backendToken, backendToken.Expiry,
	); err != nil {
		return err
	}
	sessionToken, refreshToken, err := h.sessionService.GenerateSessionTokens(session)
	if err != nil {
		return err
	}
	setCookie(c, "session", sessionToken, "/")
	setCookie(c, "refresh", refreshToken, "/refresh")
	return c.Redirect(
		http.StatusTemporaryRedirect,
		fmt.Sprintf(SessionURLFmt, req.AppID, req.State),
	)
}

// HandleSesion is triggered after the HandleCallBack redirects the user to
// iteself with a session cookie.  This will generate the access code for
// runner and redirect the user to the bifrost callback URL with the Runner
// access code.
func (h *Handler) HandleSession(c echo.Context) error {
	req, err := NewSessionRequest(c)
	if err != nil {
		return err
	}

	session, err := h.sessionService.ParseJWT(req.SessionToken, "")
	if err != nil {
		return err
	}

	code, err := h.sessionService.GenerateAccessCode(session)
	if err != nil {
		return err
	}

	redirectURL := h.cloudURL.JoinPath(
		fmt.Sprintf(BifrostCallbackURLFmt, req.AppID),
	)

	q := redirectURL.Query()
	q.Add("app_id", req.AppID)
	q.Add("code", code)
	q.Add("state", req.State)
	redirectURL.RawQuery = q.Encode()

	return c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())

}

// HandleToken handles the token request.  This completes the OAuth2 flow of
// Runner and issues a runner token in exchange for the authorization code.
func (h *Handler) HandleToken(c echo.Context) error {
	req, err := NewTokenRequest(c)
	if err != nil {
		return err
	}

	if !h.runner.IsValidClient(req.ClientID, req.ClientSecret) {
		return httperror.ErrUnauthorized(fmt.Errorf("invalid client credentials"))
	}

	session, err := h.sessionService.GetByAccessCode(req.Code)
	if err != nil {
		return err
	}

	token, err := h.sessionService.GenerateOauthToken(session)
	if err != nil {
		return err
	}

	res := NewTokenResponse(token)
	return c.JSON(http.StatusOK, res)
}

// HandleRefresh is the Refresh token endpoint.  This refreshes the underlying
// access token and issues a new token.
func (h *Handler) HandleRefresh(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := NewRefreshRequest(c)
	if err != nil {
		return err
	}

	if !h.runner.IsValidClient(req.ClientID, req.ClientSecret) {
		return httperror.ErrUnauthorized(fmt.Errorf("invalid client credentials"))
	}

	session, err := h.sessionService.ParseJWT(req.RefreshToken, "")
	if err != nil {
		return err
	}

	// Refresht the backend token.
	backendToken := new(oauth2.Token)
	if err := session.DeserializeBackendToken(backendToken); err != nil {
		return err
	}

	backendToken, err = h.backend.RefreshToken(ctx, req.AppID, backendToken.RefreshToken)
	if err != nil {
		return err
	}

	session.ExpiresAt = backendToken.Expiry.Unix()

	err = h.sessionService.SetBackendToken(session, backendToken, backendToken.Expiry)
	if err != nil {
		return err
	}

	// Generate new Runner token.
	token, err := h.sessionService.GenerateOauthToken(session)
	if err != nil {
		return err
	}

	res := NewTokenResponse(token)

	return c.JSON(http.StatusOK, res)
}

// HandleUser handles the user request.  This fetches the user from the underlying
// OAuth backend.
func (h *Handler) HandleUser(c echo.Context) error {
	req, err := NewUserRequest(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()

	session, err := h.sessionService.ParseJWT(req.AccessToken, "")
	if err != nil {
		return err
	}

	backendToken := new(oauth2.Token)
	if err := session.DeserializeBackendToken(backendToken); err != nil {
		return err
	}

	user, err := h.backend.GetUser(ctx, req.AppID, backendToken)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, &UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.Name,
		UserName: user.Login,
	})
}
func setCookie(c echo.Context, name, value, path string) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})
}
