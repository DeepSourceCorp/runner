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

	"github.com/deepsourcecorp/runner/provider/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebhookService_Process is an e2e happy path test for the webhook service.
func TestWebhookService_Process(t *testing.T) {
	body := []byte("test-body")

	runner := &common.Runner{
		ID:            "test-runner-id",
		WebhookSecret: "runner-webhook-secret",
	}

	app := &App{
		WebhookSecret: "app-webhook-secret",
	}

	// this server will act as the cloud server
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rBody, _ := io.ReadAll(r.Body)

			assert.Equal(t, "/services/webhooks/github/", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, rBody, body)
			assert.Equal(t, app.ID, r.Header.Get(common.HeaderAppID))
			assert.Equal(t, runner.ID, r.Header.Get(common.HeaderRunnerID))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "sha256=3b91deee7610a48e3bcdddd420a5bbb8ca960b7cf4c547a9eb5017ac116216c3", r.Header.Get(common.HeaderRunnerSignature))
			assert.Equal(t, r.ContentLength, int64(len(body)))

			w.WriteHeader(http.StatusOK)
		}),
	)
	defer server.Close()
	serverURL, _ := url.Parse(server.URL)

	wr := &WebhookRequest{
		AppID:       "test-app-id",
		HTTPRequest: httptest.NewRequest(http.MethodGet, "https://example.com", bytes.NewReader(body)),
		Signature:   "sha256=825e0c233e2943e5eeffe9be54ed00a1c178c4b9457337cb8abf10a61645e347",
	}

	service := NewService(&ServiceOpts{
		Runner: runner,
		Apps: map[string]*App{
			"test-app-id": app,
		},
		DeepSource: &common.DeepSource{
			Host: *serverURL,
		},
		Client: http.DefaultClient,
	})

	_, err := service.ForwardWebhook(wr)

	assert.NoError(t, err)
}

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

	httpRequest := httptest.NewRequest(http.MethodGet, "https://test.com/apps/test-app-id/api/user", bytes.NewReader(body))
	httpRequest.Header.Set(common.HeaderAuthorization, "original-authorization") // This should be removed
	httpRequest.Header.Set(common.HeaderAccept, "original-accept")               // This should be retained

	request := &APIRequest{
		AppID:          "test-app-id",
		InstallationID: "test-installation-id",
		HTTPRequest:    httpRequest,
	}

	service := NewService(&ServiceOpts{
		Runner:     &common.Runner{},
		DeepSource: &common.DeepSource{},
		Apps:       map[string]*App{"test-app-id": app},
		Client:     http.DefaultClient,
	})
	res, err := service.ForwardAPI(request)
	require.NoError(t, err)
	defer res.Body.Close()

	body, _ = io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, githubBody, body)
}
