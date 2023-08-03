package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/deepsourcecorp/runner/auth/model"
	"golang.org/x/oauth2"
)

const (
	ProviderGitlab = "gitlab"

	GitlabURLAuthorize = "/oauth/authorize"
	GitlabURLToken     = "/oauth/token"
	GitlabURLUser      = "/api/v4/user"
)

type Gitlab struct {
	config *oauth2.Config
	client *http.Client

	apiHost url.URL
}

func NewGitlab(app *App) (IBackend, error) {
	return &Gitlab{
		config: &oauth2.Config{
			ClientID:     app.ClientID,
			ClientSecret: app.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:   app.AuthHost.JoinPath(GitlabURLAuthorize).String(),
				TokenURL:  app.AuthHost.JoinPath(GitlabURLToken).String(),
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes:      []string{"api", "read_user", "read_repository"},
			RedirectURL: app.RedirectURL.String(),
		},
		client:  &http.Client{},
		apiHost: app.APIHost,
	}, nil
}

type GitlabUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	UserName string `json:"username"`
}

func (g *Gitlab) AuthorizationURL(state string, scopes []string) (string, error) {
	g.config.Scopes = append(g.config.Scopes, scopes...)
	return g.config.AuthCodeURL(state), nil
}

func (g *Gitlab) GetToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.config.Exchange(ctx, code)
}

func (g *Gitlab) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := new(oauth2.Token)
	token.RefreshToken = refreshToken
	token.Expiry = time.Now()

	ts := g.config.TokenSource(ctx, token)
	return ts.Token()
}

func (g *Gitlab) GetUser(ctx context.Context, token *oauth2.Token) (*model.User, error) {
	userURL := g.apiHost.JoinPath(GitlabURLUser).String()

	req, err := http.NewRequest(http.MethodGet, userURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("gitlab: failed to get user: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gitlab: failed to get user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gitlab: failed to get user: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gitlab: failed to get user: status=%s", resp.Status)
	}

	var u = new(GitlabUser)
	if err := json.Unmarshal(body, u); err != nil {
		return nil, fmt.Errorf("gitlab: failed to get user: %w", err)
	}

	return &model.User{
		ID:       fmt.Sprintf("%d", u.ID),
		Name:     u.Name,
		Email:    u.Email,
		Login:    u.UserName,
		Provider: ProviderGitlab,
	}, nil
}
