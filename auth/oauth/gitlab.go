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

	"github.com/deepsourcecorp/runner/auth/common"
	"golang.org/x/exp/slog"
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

	authHost *url.URL
}

func NewGitlab(app *App) Provider {
	return &Gitlab{
		config: &oauth2.Config{
			ClientID:     app.ClientID,
			ClientSecret: app.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  app.AuthHost.JoinPath(GitlabURLAuthorize).String(),
				TokenURL: app.AuthHost.JoinPath(GitlabURLToken).String(),
			},
			RedirectURL: app.RedirectURL.String(),
			Scopes:      []string{"openid", "api", "read_user", "email", "profile"},
		},
		client:   &http.Client{},
		authHost: app.AuthHost,
	}
}

func (g *Gitlab) AuthorizationURL(state string, scopes []string) string {
	g.config.Scopes = append(g.config.Scopes, scopes...)
	return g.config.AuthCodeURL(state)
}

func (g *Gitlab) GetToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	return token, nil
}

func (g *Gitlab) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := new(oauth2.Token)
	token.RefreshToken = refreshToken
	token.Expiry = time.Now().Add(-time.Hour)

	tokenSource := g.config.TokenSource(ctx, token)
	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	return token, nil
}

type GitlabUserReponse struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

func (r *GitlabUserReponse) ToModel() *common.User {
	return &common.User{
		ID:       strconv.Itoa(r.ID),
		Email:    r.Email,
		Login:    r.Username,
		Name:     r.Name,
		Provider: ProviderGitlab,
	}
}

func (g *Gitlab) GetUser(ctx context.Context, token *oauth2.Token) (*common.User, error) {
	url := g.authHost.JoinPath(GitlabURLUser).String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		slog.Error("failed to create request", slog.Any("err", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	res, err := g.client.Do(req)
	if err != nil {
		slog.Error("failed to get user", slog.Any("err", err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		slog.Error("non 2xx response from GitLab for /userinfo", slog.Any("code", res.StatusCode))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("failed to read response body", slog.Any("err", err))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	u := &GitlabUserReponse{}
	if err := json.Unmarshal(body, &u); err != nil {
		slog.Error("failed to unmarshal response body", slog.Any("err", err))
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	user := u.ToModel()
	user.Provider = ProviderGitlab
	return user, nil
}
