package oauth

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGithub_GetToken(t *testing.T) {
	clientID := "client_id_1"
	clientSecret := "client_secret_1"
	code := "code_1"
	appID := "app_id_1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			t.Fatal(err)
			return
		}
		if r.Header.Get("Authorization") != "Basic "+base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)) {
			t.Error("Github.GetToken() failed, expected Authorization header to be 'Basic base64(clientID:clientSecret)'")
			return
		}
		if r.FormValue("code") != code {
			t.Error("Github.GetToken() failed, expected code to be 'code_1'")
			return
		}
		if !strings.HasSuffix(r.FormValue("redirect_uri"), "/apps/app_id_1/oauth2/callback") {
			t.Error("Github.GetToken() failed, expected redirect_uri to be '/apps/app_id_1/oauth2/callback', got", r.FormValue("redirect_uri"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"token","token_type":"bearer","scope":"repo,gist"}`))
	}))
	defer server.Close()
	serverURL, _ := url.Parse(server.URL)
	redirectURL, _ := url.Parse("http://localhost:8080/apps/app_id_1/oauth2/callback")
	app := &App{
		ID:           appID,
		AuthHost:     *serverURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  *redirectURL,
	}
	github, err := NewGithub(app)
	if err != nil {
		t.Fatal(err)
		return
	}

	token, err := github.GetToken(context.Background(), code)
	if err != nil {
		t.Fatal(err)
		return
	}

	if token.AccessToken != "token" {
		t.Error("Github.GetToken() failed, expected token to be 'token'")
		return
	}
}
