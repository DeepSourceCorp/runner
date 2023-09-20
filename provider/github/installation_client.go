package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/deepsourcecorp/runner/proxyutil"
)

var (
	GithubURLAccessTokenFmt = "/app/installations/%s/access_tokens"

	ErrAppNotFound = fmt.Errorf("app not found")
)

const (
	HeaderValueGithubAccept = "application/vnd.github+json"
)

type InstallationClient struct {
	app            *App
	installationID string

	client *http.Client
}

func NewInstallationClient(app *App, installationID string, client *http.Client) *InstallationClient {
	return &InstallationClient{
		app:            app,
		installationID: installationID,
		client:         client,
	}
}

type AccessTokenResponse struct {
	Token string `json:"token"`
}

// AccessToken returns the Installation Access Token for the given installation ID
// and app.  This token is used to authenticate requests to the GitHub API on behalf of an
// installation.
// (https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-an-installation-access-token-for-a-github-app)
func (c *InstallationClient) AccessToken() (string, error) {
	accessTokenURL := c.app.APIHost.JoinPath(fmt.Sprintf(GithubURLAccessTokenFmt, c.installationID)).String()

	jwtToken, err := c.app.JWT()
	if err != nil {
		return "", fmt.Errorf("failed to generate jwt token: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, accessTokenURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	req.Header.Set("Accept", HeaderValueGithubAccept)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request for access token failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var accessTokenResponse AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&accessTokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode access token response: %w", err)
	}

	return accessTokenResponse.Token, nil
}

// ProxyURL returns the URL to proxy the request.  When DeepSource Cloud sends a
// request to GitHub via the Runner, it is prefixed with "/apps/:app_id/api".
// This method strips this prefix and generates the actual GitHub API URL.
func (c *InstallationClient) ProxyURL(path string) string {
	prefixToRemove := fmt.Sprintf("/apps/%s/api/", c.app.ID)
	return c.app.APIHost.JoinPath(strings.TrimPrefix(path, prefixToRemove)).String()
}

func (c *InstallationClient) Proxy(in *http.Request, accessToken string) (*http.Response, error) {
	targetURL := c.ProxyURL(in.URL.Path)
	req, err := http.NewRequest(in.Method, targetURL, in.Body)
	if err != nil {
		return nil, err
	}

	proxyutil.CopyHeader(req.Header, in.Header)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	return c.client.Do(req)
}
