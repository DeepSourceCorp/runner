package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/exp/slog"
)

var (
	GithubURLAccessTokenFmt = "/app/installations/%s/access_tokens"

	ErrAppNotFound = fmt.Errorf("app not found")
)

const (
	HeaderValueGithubAccept = "application/vnd.github+json"
)

type APIProxyFactory struct {
	apps   map[string]*App
	client *http.Client
}

func NewAPIProxyFactory(apps map[string]*App, client *http.Client) *APIProxyFactory {
	return &APIProxyFactory{
		apps:   apps,
		client: client,
	}
}

func (f *APIProxyFactory) NewProxy(appID string, installationID string) (*APIProxy, error) {
	app := f.apps[appID]
	if app == nil {
		return nil, ErrAppNotFound
	}
	return &APIProxy{
		client:         f.client,
		app:            app,
		installationID: installationID,
	}, nil
}

type APIProxy struct {
	client         *http.Client
	app            *App
	installationID string
}

// GenerateJWT generates a signed JWT token for the Github App.
func (c *APIProxy) GenerateJWT() (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": jwt.TimeFunc().Add(-1 * time.Minute).Unix(),
		"exp": jwt.TimeFunc().Add(10 * time.Minute).Unix(),
		"iss": c.app.AppID,
	}).SignedString(c.app.PrivateKey)
}

// GenerateAccessToken generates a short lived access token for the Github App
// using the JWT token.
func (c *APIProxy) GenerateAccessToken(requestToken string) (string, error) {
	tokenURL := c.accessTokenURL()
	req, err := http.NewRequest(
		http.MethodPost,
		tokenURL,
		http.NoBody,
	)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to create request for access token: %v", err))
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", requestToken))
	req.Header.Set("Accept", HeaderValueGithubAccept)

	resp, err := c.client.Do(req)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to generate access token: %v", err))
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		slog.Error(fmt.Sprintf("failed to generate access token, received: %d", resp.StatusCode))
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respBody struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}

	if respBody.Token == "" {
		return "", fmt.Errorf("empty access token")
	}

	return respBody.Token, nil
}

func (p *APIProxy) accessTokenURL() string {
	return p.app.APIHost.JoinPath(fmt.Sprintf("/app/installations/%s/access_tokens", p.installationID)).String()
}

// ProxyURL generates the Github API URL by removing the Runner API prefix from
// the request URL.
func (c *APIProxy) ProxyURL(path string) string {
	prefixToRemove := fmt.Sprintf("/apps/%s/api/", c.app.ID)
	return c.app.APIHost.JoinPath(strings.TrimPrefix(path, prefixToRemove)).String()
}

// Proxy proxies the request to the Github API after adding the required
// authentication headers.
func (c *APIProxy) Proxy(req *http.Request) (*http.Response, error) {
	requestToken, err := c.GenerateJWT()
	if err != nil {
		return nil, err
	}
	accessToken, err := c.GenerateAccessToken(requestToken)
	if err != nil {
		return nil, err
	}
	u := c.ProxyURL(req.URL.Path)
	req, err = http.NewRequest(req.Method, u, req.Body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", HeaderValueGithubAccept)

	return c.client.Do(req)
}

func (c *APIProxy) InstallationURL() string {
	return c.app.BaseHost.JoinPath(fmt.Sprintf("/apps/%s/installations/new", c.app.AppSlug)).String()
}
