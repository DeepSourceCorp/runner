package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func gitlabServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload = make(map[string]interface{})

		if r.URL.Path == "/oauth/token" {
			payload = map[string]interface{}{
				"access_token":  "gitlab-access-token",
				"token_type":    "bearer",
				"expires_in":    7200,
				"refresh_token": "gitlab-refresh-token",
				"created_at":    1562909409,
			}
		}
		if r.URL.Path == "/api/v4/user" {
			payload = map[string]interface{}{
				"id":       1,
				"username": "john_smith",
				"email":    "john@xyz.com",
				"name":     "John Smith",
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(payload)
	}))
	return server
}

func gitlabProvider(server *httptest.Server) Provider {
	serverURL, _ := url.Parse(server.URL)
	redirectURL, _ := url.Parse("http://localhost:8080/apps/app-2/auth/callback")
	app := &App{
		ClientID:     "gitlab-client-id",
		ClientSecret: "gitlab-client-secret",
		RedirectURL:  redirectURL,
		AuthHost:     serverURL,
		Provider:     ProviderGitlab,
	}

	return NewGitlab(app)
}

func TestGitlab_AuthorizationURL(t *testing.T) {
	gitlabServer := gitlabServer()
	defer gitlabServer.Close()
	provider := gitlabProvider(gitlabServer)

	got := provider.AuthorizationURL("state", []string{})
	gotURL, _ := url.Parse(got)
	assert.Equal(t, "/oauth/authorize", gotURL.Path)
	assert.Equal(t, "gitlab-client-id", gotURL.Query().Get("client_id"))
	assert.Equal(t, "state", gotURL.Query().Get("state"))
	assert.Equal(t, "code", gotURL.Query().Get("response_type"))
}

func TestGitlab_GetToken(t *testing.T) {
	server := gitlabServer()
	defer server.Close()
	provider := gitlabProvider(server)

	got, err := provider.GetToken(context.Background(), "code")
	require.NoError(t, err)
	assert.NotEmpty(t, got.AccessToken)
	assert.NotEmpty(t, got.RefreshToken)
	assert.NotEmpty(t, got.Expiry)
	assert.Equal(t, "Bearer", got.Type())
}

func TestGitlab_RefreshToken(t *testing.T) {
	server := gitlabServer()
	defer server.Close()
	provider := gitlabProvider(server)

	got, err := provider.RefreshToken(context.Background(), "refresh-token")
	require.NoError(t, err)
	assert.NotEmpty(t, got.AccessToken)
	assert.NotEmpty(t, got.RefreshToken)
	assert.NotEmpty(t, got.Expiry)
	assert.Equal(t, "Bearer", got.Type())
}

func TestGitlab_GetUser(t *testing.T) {
	server := gitlabServer()
	defer server.Close()
	provider := gitlabProvider(server)

	token := &oauth2.Token{
		AccessToken: "gitlab-access-token",
		TokenType:   "Bearer",
	}

	got, err := provider.GetUser(context.Background(), token)
	require.NoError(t, err)

	assert.Equal(t, "john_smith", got.Login)
	assert.Equal(t, "john@xyz.com", got.Email)
	assert.Equal(t, "John Smith", got.Name)
	assert.Equal(t, "1", got.ID)

}
