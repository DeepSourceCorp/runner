package token

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/deepsourcecorp/runner/auth/cryptutil"
	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionAuthMiddleware(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer := cryptutil.NewSigner(privateKey)
	verifier := cryptutil.NewVerifier(&privateKey.PublicKey)
	service := NewService(signer, verifier)
	middleware := SessionAuthMiddleware("runner-id", service)

	t.Run("cookie not set", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		h := middleware(func(c echo.Context) error {
			return c.HTML(200, "test")
		})
		err := h(c)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "/refresh", rec.Header().Get("Location"))
	})

	t.Run("cookie empty", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: ""})
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		h := middleware(func(c echo.Context) error {
			return c.HTML(200, "test")
		})
		err := h(c)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "/refresh", rec.Header().Get("Location"))
	})

	t.Run("cookie invalid", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: "invalid"})
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		h := middleware(func(c echo.Context) error {
			return c.HTML(200, "test")
		})
		err := h(c)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "/refresh", rec.Header().Get("Location"))
	})

	user := &model.User{
		ID:       "test-id",
		Email:    "abc@xyz.com",
		Name:     "test-name",
		Login:    "test-login",
		Provider: "test-provider",
	}

	t.Run("token expired", func(t *testing.T) {
		ExpiryAccessToken = -1 * time.Minute
		token, err := service.GenerateToken("runner-id", []string{ScopeUser}, user, ExpiryAccessToken)
		require.NoError(t, err)
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		h := middleware(func(c echo.Context) error {
			return c.HTML(200, "ok")
		})
		err = h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "/refresh", rec.Header().Get("Location"))
	})

	t.Run("valid token", func(t *testing.T) {
		ExpiryAccessToken = 10 * time.Minute
		token, err := service.GenerateToken("runner-id", []string{ScopeCodeRead}, user, ExpiryAccessToken)
		require.NoError(t, err)
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		h := middleware(func(c echo.Context) error {
			return c.HTML(200, "ok")
		})
		err = h(c)
		assert.NoError(t, err)
		assert.Equal(t, 200, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})
}
