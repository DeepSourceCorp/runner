package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/deepsourcecorp/runner/proxyutil"
	"golang.org/x/exp/slog"
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

		client: client,
	}
}

type AccessTokenResponse struct {
	Token string `json:"token"`
}

func (c *InstallationClient) AccessToken() (string, error) {
	accessTokenURL := c.app.APIHost.JoinPath(fmt.Sprintf(GithubURLAccessTokenFmt, c.installationID)).String()

	jwtToken, err := c.app.JWT()
	if err != nil {
		slog.Error("failed to generate jwt token", slog.Any("err", err))
		return "", fmt.Errorf("failed to generate jwt token: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, accessTokenURL, http.NoBody)
	if err != nil {
		slog.Error("failed to create request for access token", slog.Any("err", err))
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	req.Header.Set("Accept", HeaderValueGithubAccept)

	resp, err := c.client.Do(req)
	if err != nil {
		slog.Error("failed to generate access token", slog.Any("err", err))
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		slog.Error("failed to generate access token", slog.Any("status_code", resp.StatusCode))
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var accessTokenResponse AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&accessTokenResponse); err != nil {
		slog.Error("failed to decode access token response", slog.Any("err", err))
		return "", err
	}

	return accessTokenResponse.Token, nil
}

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
