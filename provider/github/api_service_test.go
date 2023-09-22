package github

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIService_Process(t *testing.T) {
	body := []byte("test-body")
	githubBody := []byte(`{"id": 1}`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app/installations/test-installation-id/access_tokens" {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token": "test-token"}`))
			return
		}
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/user", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "original-accept", r.Header.Get("Accept"))

		gotBody, _ := io.ReadAll(r.Body)
		assert.Equal(t, body, gotBody)
		assert.Equal(t, len(body), int(r.ContentLength))

		_, _ = w.Write([]byte(githubBody))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	serverURL, _ := url.Parse(server.URL)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	app := &App{
		ID:         "test-app-id",
		PrivateKey: privateKey,
		APIHost:    *serverURL,
	}

	appFactory := &AppFactory{
		apps: map[string]*App{"test-app-id": app},
	}

	httpRequest := httptest.NewRequest(http.MethodGet, "https://test.com/apps/test-app-id/api/user", bytes.NewReader(body))
	httpRequest.Header.Set(HeaderAuthorization, "original-authorization") // This should be removed
	httpRequest.Header.Set(HeaderAccept, "original-accept")               // This should be retained

	request := &APIRequest{
		AppID:          "test-app-id",
		InstallationID: "test-installation-id",
		HTTPRequest:    httpRequest,
	}

	service := NewAPIService(appFactory, http.DefaultClient)
	res, err := service.Process(request)
	require.NoError(t, err)
	defer res.Body.Close()

	body, _ = io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, githubBody, body)
}
