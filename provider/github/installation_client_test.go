package github

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallationClient_AccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app/installations/test-installation-id/access_tokens" {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token": "test-token"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	serverURL, _ := url.Parse(server.URL)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	app := &App{
		ID:         "test-app-id",
		AppSlug:    "test-app-slug",
		BaseHost:   *serverURL,
		APIHost:    *serverURL,
		PrivateKey: privateKey,
	}

	installationClient := &InstallationClient{app: app, installationID: "test-installation-id", client: http.DefaultClient}

	accessToken, err := installationClient.AccessToken()

	assert.NoError(t, err)
	assert.Equal(t, "test-token", accessToken)
}
