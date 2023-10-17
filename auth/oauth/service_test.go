package oauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/deepsourcecorp/runner/auth/common"
	"github.com/deepsourcecorp/runner/auth/contract"
	"github.com/deepsourcecorp/runner/auth/session"
	"github.com/deepsourcecorp/runner/auth/store/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	mockStore      = mock.NewInMemorySessionStore()
)

func getMockBackendServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]interface{}{
			"access_token":  "backend-access-token",
			"refresh_token": "backend-refresh-token",
			"expires_in":    3600,
			"token_type":    "bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(payload)
	}))
}

func service() *Service {
	mockBackenServer := getMockBackendServer()

	redirectURL, _ := url.Parse("http://redirecturl.com")
	deepsourceURL, _ := url.Parse("http://deepsourceurl.com")
	providerURL, _ := url.Parse(mockBackenServer.URL)

	runner := &common.Runner{ID: "runner-id", ClientID: "client-id", ClientSecret: "client-secret", PrivateKey: _privateKey}

	deepsource := &common.DeepSource{BaseURL: deepsourceURL}

	sessionService := session.NewService(runner, deepsource, mockStore)

	apps := map[string]*App{
		"app-id": {
			ClientID:     "backend-client-id1",
			ClientSecret: "backend-client-secret1",
			RedirectURL:  redirectURL,
			AuthHost:     providerURL,
			APIHost:      providerURL,
			Provider:     ProviderGithub,
		},
		"invalid-provider-app-id": {
			ClientID:     "backend-client-id2",
			ClientSecret: "backend-client-secret2",
			RedirectURL:  redirectURL,
			AuthHost:     providerURL,
			APIHost:      providerURL,
			Provider:     "invalid-provider",
		},
	}
	return NewService(apps, sessionService)
}

func TestOAuthService_GetAuthorizationURL(t *testing.T) {
	oauthService := service()

	t.Run("should return authorization URL", func(t *testing.T) {
		req := &contract.AuthorizationRequest{
			AppID:    "app-id",
			ClientID: "client-id",
			State:    "state",
			Scopes:   []string{"scope1", "scope2"},
		}
		got, err := oauthService.GetAuthorizationURL(req)
		require.NoError(t, err)

		gotURL, err := url.Parse(got)
		require.NoError(t, err)

		assert.Equal(t, "http", gotURL.Scheme)
		assert.Equal(t, "/login/oauth/authorize", gotURL.Path)

		query := gotURL.Query()
		assert.Equal(t, "backend-client-id1", query.Get("client_id"))
		assert.Equal(t, "state", query.Get("state"))
		assert.Equal(t, "read:user user:email scope1 scope2", query.Get("scope"))
		assert.Equal(t, "http://redirecturl.com", query.Get("redirect_uri"))
	})

	t.Run("should return error if client id is invalid", func(t *testing.T) {
		req := &contract.AuthorizationRequest{
			AppID:    "app-id",
			ClientID: "invalid-client-id",
			State:    "state",
			Scopes:   []string{"scope1", "scope2"},
		}
		got, err := oauthService.GetAuthorizationURL(req)
		require.Error(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("should return error if app id is invalid", func(t *testing.T) {
		req := &contract.AuthorizationRequest{
			AppID:    "invalid-provider-app-id",
			ClientID: "client-id",
			State:    "state",
			Scopes:   []string{"scope1", "scope2"},
		}
		got, err := oauthService.GetAuthorizationURL(req)
		require.Error(t, err)
		assert.Equal(t, "", got)
	})
}

func TestOAuthService_CreateSession(t *testing.T) {
	oauthService := service()

	t.Run("should return session", func(t *testing.T) {
		req := &contract.CallbackRequest{
			AppID: "app-id",
			Code:  "code",
		}

		got, err := oauthService.CreateSession(context.Background(), req)

		require.NoError(t, err, errors.Unwrap(err))
		assert.NotEmpty(t, got.ID)
		assert.NotEmpty(t, got.BackendToken)

	})

	t.Run("should return error if app id is invalid", func(t *testing.T) {
		req := &contract.CallbackRequest{
			AppID: "invalid-provider-app-id",
			Code:  "code",
		}

		got, err := oauthService.CreateSession(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, got)
	})
}
