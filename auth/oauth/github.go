package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

const (
	ProviderGithub = "github"

	// Github API endpoints
	GithubURLAuthorize = "/login/oauth/authorize"
	GithubURLToken     = "/login/oauth/access_token"
	GithubURLUser      = "/user"
	GithubURLEmails    = "/user/emails"
)

type Github struct {
	config *oauth2.Config
	client *http.Client

	authHost url.URL
	apiHost  url.URL
}

func NewGithub(app *App) (IBackend, error) {
	return &Github{
		config: &oauth2.Config{
			ClientID:     app.ClientID,
			ClientSecret: app.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  app.AuthHost.JoinPath(GithubURLAuthorize).String(),
				TokenURL: app.AuthHost.JoinPath(GithubURLToken).String(),
			},
			RedirectURL: app.RedirectURL.String(),
			Scopes:      []string{},
		},
		client:   &http.Client{},
		authHost: app.AuthHost,
		apiHost:  app.APIHost,
	}, nil
}

func (g *Github) AuthorizationURL(state string, scopes []string) (string, error) {
	g.config.Scopes = append(g.config.Scopes, scopes...)
	return g.config.AuthCodeURL(state), nil
}

func (g *Github) GetToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.config.Exchange(ctx, code)
}

func (g *Github) RefreshToken(_ context.Context, refreshToken string) (*oauth2.Token, error) {
	payload := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		GrantType    string `json:"grant_type"`
		RefreshToken string `json:"refresh_token"`
	}{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("oauth(gh): failed to marshal payload while refreshing token", slog.Any("err", err))
		return nil, err
	}

	refreshTokenURL := g.config.Endpoint.TokenURL

	req, err := http.NewRequest("POST", refreshTokenURL, bytes.NewReader(body))
	if err != nil {
		slog.Error("oauth(gh): failed to create request while refreshing token", slog.Any("err", err))
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		slog.Error("oauth(gh): failed to do request while refreshing token", slog.Any("err", err))
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("oauth(gh): non-200 status code while refreshing token", slog.Any("status", resp.StatusCode))
		return nil, errors.New("error refreshing token")
	}

	defer resp.Body.Close()

	var response struct {
		AccessToken           string `json:"access_token"`
		ExpiresIn             int    `json:"expires_in"`
		RefreshToken          string `json:"refresh_token"`
		RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
		Scope                 string `json:"scope"`
		TokenType             string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		slog.Error("oauth(gh): failed to decode response while refreshing token", slog.Any("err", err))
		return nil, err
	}

	if response.AccessToken == "" {
		slog.Error("oauth(gh): no access token while refreshing token", slog.Any("response", response))
		return nil, errors.New("no access token")
	}

	return &oauth2.Token{
		AccessToken: response.AccessToken,
		// Since Github only gives the seconds until expiry, subtracting ~10
		// second from the expiry.  This is to account for any latency between
		//Runner and Github.
		// Note: This is weird.
		Expiry:       time.Now().Add(time.Duration(response.ExpiresIn-10) * time.Second),
		RefreshToken: response.RefreshToken,
		TokenType:    response.TokenType,
	}, nil
}

func (g *Github) GetUser(ctx context.Context, token *oauth2.Token) (*User, error) {
	userURL := g.apiHost.JoinPath(GithubURLUser)
	req, err := http.NewRequest("GET", userURL.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		slog.Error("error getting user", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	slog.Debug("github user", "body", string(body))

	var u struct {
		Email string `json:"email"`
		Login string `json:"login"`
		Name  string `json:"name"`
		ID    int    `json:"id"`
	}

	if err := json.Unmarshal(body, &u); err != nil {
		slog.Error("error unmarshalling user", "err", err)
		return nil, err
	}

	if u.Email == "" {
		slog.Info("no email found, getting primary email")
		email, err := g.getPrimaryEmail(ctx, token)
		if err != nil {
			return nil, err
		}
		u.Email = email
	}

	return &User{
		ID:    strconv.Itoa(u.ID),
		Email: u.Email,
		Login: u.Login,
		Name:  u.Name,
	}, nil
}

func (g *Github) getPrimaryEmail(_ context.Context, token *oauth2.Token) (string, error) {
	emailURL := g.apiHost.JoinPath(GithubURLEmails)
	req, err := http.NewRequest("GET", emailURL.String(), http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	slog.Debug("email response body", string(body))
	if err != nil {
		slog.Error("error reading body", "err", err)
		return "", err
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
		slog.Error("error unmarshalling emails", "err", err)
		return "", err
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}
	return "", errors.New("no primary email found")
}
