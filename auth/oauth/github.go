package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/deepsourcecorp/runner/auth/model"
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
			Scopes:      []string{"read:user", "user:email"},
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

func (g *Github) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := new(oauth2.Token)
	token.RefreshToken = refreshToken
	token.Expiry = time.Now()
	return g.config.TokenSource(ctx, token).Token()
}

func (g *Github) GetUser(ctx context.Context, token *oauth2.Token) (*model.User, error) {
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

	return &model.User{
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
