package oauth

import (
	"context"
	"encoding/json"
	"fmt"
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

	host, _ := url.Parse("https://github.com")
	redirect, _ := url.Parse("http://example.com/apps/app2/auth/callback")

	app := &App{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthHost:     *host,
		RedirectURL:  *redirect,
	}

	github, err := NewGithub(app)
	assert.NoError(t, err)

	authroization := github.AuthorizationURL("state", []string{})

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

func TestGithub_GetUser(t *testing.T) {
	response := GithubUserResponse{
		Login: "login",
		Email: "email",
		Name:  "name",
		ID:    1,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))

	host, _ := url.Parse(server.URL)

	github := &Github{
		apiHost: *host,
		client:  http.DefaultClient,
	}

	token := &oauth2.Token{
		AccessToken: "access-token",
	}

	user, err := github.GetUser(context.Background(), token)
	assert.NoError(t, err)

	assert.Equal(t, response.Login, user.Login)
	assert.Equal(t, response.Email, user.Email)
	assert.Equal(t, response.Name, user.Name)
	assert.Equal(t, fmt.Sprintf("%d", response.ID), user.ID)
}

func TestGithub_GetUserEmails(t *testing.T) {
	response := []GithubEmail{
		{
			Email:    "email1",
			Verified: false,
			Primary:  false,
		},
		{
			Email:    "email2",
			Verified: true,
			Primary:  true,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))

	host, _ := url.Parse(server.URL)

	github := &Github{
		apiHost: *host,
		client:  http.DefaultClient,
	}

	token := &oauth2.Token{
		AccessToken: "access-token",
	}

	primary, err := github.getPrimaryEmail(context.Background(), token)
	assert.NoError(t, err)

	assert.Equal(t, response[1].Email, primary)

}
