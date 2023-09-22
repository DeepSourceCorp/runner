package github

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func getGitHubStub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app/installations/test-installation-id/access_tokens" {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token": "test-token"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": 1}`))
	}))
}

func TestHandler_HandleAPI(t *testing.T) {
	githubServerStub := getGitHubStub()
	defer githubServerStub.Close()
	githubServerURL, _ := url.Parse(githubServerStub.URL)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	app := &App{
		ID:         "test-app-id",
		PrivateKey: privateKey,
		APIHost:    *githubServerURL,
	}

	appFactory := &AppFactory{
		apps: map[string]*App{"test-app-id": app},
	}

	t.Run("valid api request", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "https://test.com/apps/test-app-id/api/user", nil)
		req.Header.Set(HeaderInstallationID, "test-installation-id")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("app_id")
		c.SetParamValues("test-app-id")

		service := NewAPIService(appFactory, http.DefaultClient)

		handler := &Handler{
			apiService: service,
			httpClient: http.DefaultClient,
		}
		err := handler.HandleAPI(c)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `{"id": 1}`, rec.Body.String())
	})

	t.Run("invalid app id", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "https://test.com/apps/test-app-id/api/user", nil)
		req.Header.Set(HeaderInstallationID, "test-installation-id")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("app_id")
		c.SetParamValues("invalid-app-id")

		service := NewAPIService(appFactory, http.DefaultClient)

		handler := &Handler{
			apiService: service,
			httpClient: http.DefaultClient,
		}
		err := handler.HandleAPI(c)
		assert.Error(t, err)
	})

	t.Run("invalid installation id", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "https://test.com/apps/test-app-id/api/user", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("app_id")
		c.SetParamValues("test-app-id")

		service := NewAPIService(appFactory, http.DefaultClient)

		handler := &Handler{
			apiService: service,
			httpClient: http.DefaultClient,
		}
		err := handler.HandleAPI(c)
		assert.Error(t, err)
	})
}
