package github

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/provider/model"
	"github.com/deepsourcecorp/runner/testutil"
	"github.com/stretchr/testify/assert"
)

// TestWebhookService_Process is an e2e happy path test for the webhook service.
func TestWebhookService_Process(t *testing.T) {
	body := []byte("test-body")

	runner := &model.Runner{
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
			assert.Equal(t, app.ID, r.Header.Get(HeaderAppID))
			assert.Equal(t, runner.ID, r.Header.Get(HeaderRunnerID))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "sha256=3b91deee7610a48e3bcdddd420a5bbb8ca960b7cf4c547a9eb5017ac116216c3", r.Header.Get(HeaderRunnerSignature))
			assert.Equal(t, r.ContentLength, int64(len(body)))

			w.WriteHeader(http.StatusOK)
		}),
	)
	serverURL, _ := url.Parse(server.URL)

	wr := &WebhookRequest{
		AppID:       "test-app-id",
		HTTPRequest: httptest.NewRequest(http.MethodGet, "https://example.com", bytes.NewReader(body)),
		Signature:   "sha256=825e0c233e2943e5eeffe9be54ed00a1c178c4b9457337cb8abf10a61645e347",
	}

	service := NewWebhookService(
		&AppFactory{
			apps: map[string]*App{
				"test-app-id": app,
			},
		}, runner, &model.DeepSource{
			Host: *serverURL,
		}, http.DefaultClient)

	_, err := service.Process(wr)

	assert.NoError(t, err)
}

func getWebhookService() *WebhookService {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	serverURL, _ := url.Parse(server.URL)

	appFactory := &AppFactory{
		apps: map[string]*App{
			"test-app-id": {
				WebhookSecret: "app-webhook-secret",
			},
		},
	}

	runner := &model.Runner{WebhookSecret: "runner-webhook-secret"}
	deepsource := &model.DeepSource{Host: *serverURL}

	return &WebhookService{
		appFactory: appFactory,
		runner:     runner,
		deepsource: deepsource,
	}
}

func TestWebhookService_Process_InvalidApp(t *testing.T) {
	service := getWebhookService()

	wr := &WebhookRequest{
		AppID:       "invalid-app-id",
		HTTPRequest: httptest.NewRequest(http.MethodGet, "https://example.com", nil),
	}

	_, err := service.Process(wr)

	assert.Equal(t, err, httperror.ErrAppInvalid(nil))
}

func TestWebhookService_Process_InvalidSignature(t *testing.T) {
	service := getWebhookService()

	wr := &WebhookRequest{
		AppID:       "test-app-id",
		HTTPRequest: httptest.NewRequest(http.MethodGet, "https://example.com", nil),
		Signature:   "invalid-signature",
	}

	_, err := service.Process(wr)

	assert.Equal(t, 401, err.(*httperror.Error).Code)
	assert.Equal(t, "unauthorized", err.(*httperror.Error).Message)
}

func TestWebhookService_Process_InvalidBody(t *testing.T) {
	service := getWebhookService()

	wr := &WebhookRequest{
		AppID:       "test-app-id",
		HTTPRequest: httptest.NewRequest(http.MethodGet, "https://example.com", &testutil.MockReader{Err: assert.AnError}),
	}

	_, err := service.Process(wr)

	assert.Equal(t, 500, err.(*httperror.Error).Code)
	assert.Equal(t, "unknown error", err.(*httperror.Error).Message)

}
