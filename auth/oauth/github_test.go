package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestGithub_AuthorizationURL(t *testing.T) {
	clientID := "client-id"
	clientSecret := "client-secret"

	host, _ := url.Parse("https://github.com.com")
	redirect, _ := url.Parse("http://example.com/apps/app2/auth/callback")

	app := &App{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthHost:     *host,
		RedirectURL:  *redirect,
	}

	github, err := NewGithub(app)
	assert.NoError(t, err)

	authroization, err := github.AuthorizationURL("state", []string{})
	assert.NoError(t, err)

	u, _ := url.Parse(authroization)
	q := u.Query()
	assert.Equal(t, clientID, q.Get("client_id"))
	assert.Equal(t, "state", q.Get("state"))
	assert.Equal(t, "read:user user:email", q.Get("scope"))
	assert.Equal(t, "code", q.Get("response_type"))
	assert.Equal(t, redirect.String(), q.Get("redirect_uri"))
}

func TestGithub_GetToken(t *testing.T) {
	clientID := "client-id"
	clientSecret := "client-secret"

	code := "code"
	redirect, _ := url.Parse("http://localhost:8080/apps/app2/auth/callback")

	body := []byte(`{
		"access_token": "XXXXXXXXXXXXXXXXXXXX",
		"token_type": "bearer",
		"expires_in": 7200,
		"refresh_token": "XXXXXXXXXXXXXXXXXXXX",
		"created_at": 1607635748
	   }`)

	expected := &oauth2.Token{}
	_ = json.Unmarshal(body, expected)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		err := r.ParseForm()
		assert.NoError(t, err)
		assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
		assert.Equal(t, clientID, r.Form.Get("client_id"))
		assert.Equal(t, clientSecret, r.Form.Get("client_secret"))
		assert.Equal(t, code, r.Form.Get("code"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))

	host, _ := url.Parse(server.URL)

	app := &App{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthHost:     *host,
		RedirectURL:  *redirect,
	}

	github, err := NewGithub(app)
	assert.NoError(t, err)

	token, err := github.GetToken(context.Background(), code)
	assert.NoError(t, err)
	assert.Equal(t, expected.AccessToken, token.AccessToken)
	assert.Equal(t, expected.RefreshToken, token.RefreshToken)
	assert.Equal(t, expected.TokenType, token.TokenType)
}

func TestGithub_RefreshToken(t *testing.T) {
	clientID := "client-id"
	clientSecret := "client-secret"

	refreshToken := "refresh-token"
	redirect, _ := url.Parse("http://localhost:8080/apps/app2/auth/callback")

	body := []byte(`{
		"access_token": "XXXXXXXXXXXXXXXXXXXX",
		"token_type": "bearer",
		"expires_in": 7200,
		"refresh_token": "XXXXXXXXXXXXXXXXXXXX",
		"created_at": 1607635748
	   }`)

	expected := &oauth2.Token{}
	_ = json.Unmarshal(body, expected)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		err := r.ParseForm()
		assert.NoError(t, err)
		assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
		assert.Equal(t, clientID, r.Form.Get("client_id"))
		assert.Equal(t, clientSecret, r.Form.Get("client_secret"))
		assert.Equal(t, refreshToken, r.Form.Get("refresh_token"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))

	_ = server

	host, _ := url.Parse(server.URL)

	app := &App{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthHost:     *host,
		RedirectURL:  *redirect,
	}

	github, err := NewGithub(app)
	assert.NoError(t, err)

	token, err := github.RefreshToken(context.Background(), refreshToken)

	assert.NoError(t, err)
	assert.Equal(t, expected.AccessToken, token.AccessToken)
	assert.Equal(t, expected.RefreshToken, token.RefreshToken)
	assert.Equal(t, expected.TokenType, token.TokenType)
}
