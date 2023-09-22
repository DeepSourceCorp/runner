package sync

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/deepsourcecorp/runner/auth/jwtutil"
	"github.com/stretchr/testify/assert"
)

func TestSyncer_Sync(t *testing.T) {
	apps := []App{
		{
			ID:       "app-id",
			Name:     "app-name",
			Provider: "github",
		},
	}

	runner := &Runner{
		ID:            "runner-id",
		ClientID:      "client-id",
		ClientSecret:  "client-secret",
		WebhookSecret: "webhook-secret",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/runner/", r.URL.Path)

		payload := &Payload{}
		err := json.NewDecoder(r.Body).Decode(payload)
		assert.NoError(t, err)

		assert.Equal(t, runner.ID, r.Header.Get("X-Runner-ID"))
		assert.Equal(t, runner.ID, payload.RunnerID)
		assert.Equal(t, runner.Host.String(), payload.BaseURL)
		assert.Equal(t, runner.ClientID, payload.ClientID)
		assert.Equal(t, runner.ClientSecret, payload.ClientSecret)
		assert.Equal(t, runner.WebhookSecret, payload.WebhookSecret)
		assert.Equal(t, apps, payload.Apps)
	}))

	runnerHost, _ := url.Parse("https://deepsource.io")
	runner.Host = *runnerHost

	deepsourceHost, _ := url.Parse(server.URL)
	deepsource := &DeepSource{
		Host: *deepsourceHost,
	}

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer := jwtutil.NewSigner(privateKey)

	syncer := New(deepsource, runner, apps, signer, nil)
	err := syncer.Sync()
	assert.NoError(t, err)
	server.Close()

}
