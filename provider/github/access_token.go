package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	GithubURLAccessTokenFmt = "/app/installations/%s/access_tokens"
)

func GetAccessToken(app *App, installationID string, client *http.Client) (string, error) {
	accessTokenURL := app.APIHost.JoinPath(fmt.Sprintf(GithubURLAccessTokenFmt, installationID)).String()

	jwtToken, err := app.JWT()
	if err != nil {
		return "", fmt.Errorf("failed to generate jwt token: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, accessTokenURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	req.Header.Set("Accept", HeaderValueGithubAccept)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request for access token failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var accessTokenResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accessTokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode access token response: %w", err)
	}

	if accessTokenResponse.Token == "" {
		return "", fmt.Errorf("access token not found in response")
	}

	return accessTokenResponse.Token, nil
}
