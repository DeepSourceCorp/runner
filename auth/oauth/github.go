package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/deepsourcecorp/runner/model"
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

func NewGithub(app *App) Backend {
	return &Github{
		config: &oauth2.Config{
			ClientID:     app.ClientID,
			ClientSecret: app.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  app.AuthHost.JoinPath(GithubURLAuthorize).String(),
				TokenURL: app.AuthHost.JoinPath(GithubURLToken).String(),
			},
			RedirectURL: app.RedirectURL.String(),
			Scopes:      []string{"read:user", "user:email"},
		},
		client:   &http.Client{},
		authHost: app.AuthHost,
		apiHost:  app.APIHost,
	}
}

func (g *Github) AuthorizationURL(state string, scopes []string) string {
	g.config.Scopes = append(g.config.Scopes, scopes...)
	return g.config.AuthCodeURL(state)
}

func (g *Github) GetToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	return token, nil
}

func (g *Github) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := new(oauth2.Token)
	token.RefreshToken = refreshToken
	token.Expiry = time.Now().Add(-time.Hour)

	ts := g.config.TokenSource(ctx, token)
	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	return token, nil
}

type GithubUserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

func (r *GithubUserResponse) ToModel() *model.User {
	return &model.User{
		ID:    strconv.Itoa(r.ID),
		Email: r.Email,
		Login: r.Login,
		Name:  r.Name,
	}
}

func (g *Github) GetUser(ctx context.Context, token *oauth2.Token) (*model.User, error) {
	url := g.apiHost.JoinPath(GithubURLUser).String()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		http.NoBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do user request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user response body: %w", err)
	}

	u := &GithubUserResponse{}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}

	if u.Email == "" {
		email, err := g.getPrimaryEmail(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to get primary email: %w", err)
		}
		u.Email = email
	}
	user := u.ToModel()
	user.Provider = ProviderGithub
	return u.ToModel(), nil
}

type GithubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type GithubEmails []GithubEmail

func (r GithubEmails) PrimaryEmail() (string, error) {
	for _, email := range r {
		if email.Primary {
			return email.Email, nil
		}
	}
	return "", fmt.Errorf("no primary email found")
}

func (g *Github) getPrimaryEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	url := g.apiHost.JoinPath(GithubURLEmails).String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	emails := GithubEmails{}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", fmt.Errorf("failed to get unmarshal email response: %w", err)
	}

	return emails.PrimaryEmail()
}
