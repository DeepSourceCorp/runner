package oauth

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/deepsourcecorp/runner/auth/session"
	"github.com/deepsourcecorp/runner/jwtutil"
	"github.com/deepsourcecorp/runner/model"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

var (
	// now, _              = time.Parse(time.RFC3339, "2021-07-01T00:00:00Z")
	sessionStore        = session.NewMockStore()
	runnerPrivateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
)

func mockGithubBackend() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/oauth/access_token" {
			token := map[string]interface{}{
				"access_token":  "access-token",
				"token_type":    "bearer",
				"expires_in":    3600,
				"refresh_token": "refresh-token",
			}

			body, _ := json.Marshal(token)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		}
		if r.URL.Path == "/user" {
			user := map[string]interface{}{
				"id":    1,
				"login": "deepsource",
				"email": "abc@example.com",
				"name":  "Duck Norris",
			}
			body, _ := json.Marshal(user)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		}
	}))
}

func handler() *Handler {
	mockBackendURL, _ := url.Parse(mockGithubBackend().URL)
	cloudURL, _ := url.Parse("https://deepsource.io")

	runner := &model.Runner{
		ClientID:     "runner-client-id",
		ClientSecret: "runner-client-secret",
		PrivateKey:   runnerPrivateKey,
	}

	apps := map[string]*App{
		"app-id": {
			ID:           "app-id",
			ClientID:     "app-client-id",
			ClientSecret: "app-client-secret",
			Provider:     ProviderGithub,
			AuthHost:     *mockBackendURL,
			APIHost:      *mockBackendURL,
		},
	}

	handler, _ := NewHandler(
		runner,
		cloudURL,
		apps,
		session.NewService(runner, sessionStore))
	return handler
}

func TestHandler_HandleAuthorize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/authorize", nil)
	q := req.URL.Query()
	q.Set("client_id", "runner-client-id")
	q.Set("state", "random-state")
	req.URL.RawQuery = q.Encode()

	rec := httptest.NewRecorder()

	echo := echo.New()
	c := echo.NewContext(req, rec)
	c.SetParamNames("app_id")
	c.SetParamValues("app-id")

	err := handler().HandleAuthorize(c)
	require.NoError(t, err)

	require.Equal(t, http.StatusTemporaryRedirect, rec.Code)

	redirectURL, err := url.Parse(rec.Header().Get("Location"))
	assert.NoError(t, err)

	assert.Equal(t, "/login/oauth/authorize", redirectURL.Path)
	assert.Equal(t, "app-client-id", redirectURL.Query().Get("client_id"))
	assert.Equal(t, "random-state", redirectURL.Query().Get("state"))
	assert.Equal(t, "code", redirectURL.Query().Get("response_type"))
	assert.Equal(t, "read:user user:email ", redirectURL.Query().Get("scope"))
}

func TestHandler_HandleCallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/callback", nil)
	q := req.URL.Query()
	q.Set("code", "random-code")
	q.Set("state", "random-state")
	req.URL.RawQuery = q.Encode()

	rec := httptest.NewRecorder()

	echo := echo.New()
	c := echo.NewContext(req, rec)
	c.SetParamNames("app_id")
	c.SetParamValues("app-id")

	err := handler().HandleCallback(c)
	require.NoError(t, err, errors.Unwrap(err))

	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 2)

	assert.NotEmpty(t, cookies[0].Value)
	assert.Equal(t, "/", cookies[0].Path)
	assert.Equal(t, "session", cookies[0].Name)

	assert.NotEmpty(t, cookies[1].Value)
	assert.Equal(t, "/refresh", cookies[1].Path)
	assert.Equal(t, "refresh", cookies[1].Name)

	redirectURL, err := url.Parse(rec.Header().Get("Location"))
	assert.NoError(t, err)

	assert.Equal(t, "/apps/app-id/auth/session", redirectURL.Path)

}

func TestHandler_HandleSession(t *testing.T) {
	setupSession()

	signer := jwtutil.NewSigner(runnerPrivateKey)
	token, _ := signer.GenerateToken(
		"runner-id",
		[]string{session.ScopeCode},
		map[string]interface{}{
			session.ClaimSessionID: "session-id",
		}, time.Hour,
	)

	req := httptest.NewRequest(http.MethodGet, "/session", nil)
	q := req.URL.Query()
	q.Set("state", "random-state")
	req.URL.RawQuery = q.Encode()
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: token,
		Path:  "/",
	})

	rec := httptest.NewRecorder()

	echo := echo.New()
	c := echo.NewContext(req, rec)
	c.SetParamNames("app_id")
	c.SetParamValues("app-id")

	err := handler().HandleSession(c)
	require.NoError(t, err, errors.Unwrap(err))
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)

	redirectURL, err := url.Parse(rec.Header().Get("Location"))
	assert.NoError(t, err)

	assert.Equal(t, "/accounts/runner/apps/app-id/login/callback/bifrost/", redirectURL.Path)
	assert.NotEmpty(t, redirectURL.Query().Get("code"))
	assert.Equal(t, "random-state", redirectURL.Query().Get("state"))
}

func setupSession() *session.Session {
	backendToken := &oauth2.Token{
		AccessToken:  "backend-access-token",
		TokenType:    "bearer",
		RefreshToken: "backend-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}

	s := &session.Session{
		ID:        "session-id",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Code:      "session-code",
	}
	_ = s.SetBackendToken(backendToken)

	_ = sessionStore.Create(s)
	return s
}

func TestHandler_HandleToken(t *testing.T) {
	setupSession()

	payload := map[string]string{
		"code":          "session-code",
		"client_id":     "runner-client-id",
		"client_secret": "runner-client-secret",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetParamNames("app_id")
	c.SetParamValues("app-id")

	err := handler().HandleToken(c)

	require.NoError(t, err, errors.Unwrap(err))

	assert.Equal(t, http.StatusOK, rec.Code)

	var res *TokenResponse
	err = json.NewDecoder(rec.Body).Decode(&res)
	require.NoError(t, err, errors.Unwrap(err))

	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	assert.Equal(t, "bearer", res.TokenType)
	assert.GreaterOrEqual(t, res.ExpiresIn, 0)
}

func TestHandler_HandleRefresh(t *testing.T) {
	setupSession()

	signer := jwtutil.NewSigner(runnerPrivateKey)
	refreshToken, _ := signer.GenerateToken(
		"runner-id",
		[]string{session.ScopeCode},
		map[string]interface{}{
			session.ClaimSessionID: "session-id",
		}, time.Hour,
	)

	payload := map[string]string{
		"refresh_token": refreshToken,
		"client_id":     "runner-client-id",
		"client_secret": "runner-client-secret",
	}

	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetParamNames("app_id")
	c.SetParamValues("app-id")

	err := handler().HandleRefresh(c)

	require.NoError(t, err, errors.Unwrap(err))
	assert.Equal(t, http.StatusOK, rec.Code)

	var res *TokenResponse
	err = json.NewDecoder(rec.Body).Decode(&res)
	require.NoError(t, err, errors.Unwrap(err))

	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	assert.Equal(t, "bearer", res.TokenType)
	assert.GreaterOrEqual(t, res.ExpiresIn, 0)
}

func TestHandler_HandleUser(t *testing.T) {
	setupSession()
	signer := jwtutil.NewSigner(runnerPrivateKey)
	accessToken, _ := signer.GenerateToken(
		"runner-id",
		[]string{session.ScopeCode},
		map[string]interface{}{
			session.ClaimSessionID: "session-id",
		}, time.Hour,
	)

	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetParamNames("app_id")
	c.SetParamValues("app-id")

	err := handler().HandleUser(c)

	require.NoError(t, err, errors.Unwrap(err))
	assert.Equal(t, http.StatusOK, rec.Code)

	res := &UserResponse{}
	err = json.NewDecoder(rec.Body).Decode(res)
	require.NoError(t, err, errors.Unwrap(err))

	assert.Equal(t, "1", res.ID)
	assert.Equal(t, "Duck Norris", res.FullName)
	assert.Equal(t, "abc@example.com", res.Email)
	assert.Equal(t, "deepsource", res.UserName)
}
