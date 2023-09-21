package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Signer interface {
	GenerateToken(issuer string, scope []string, claims map[string]interface{}, expiry time.Duration) (string, error)
}

type App struct {
	ID       string `json:"app_id"`
	Name     string `json:"name"`
	Provider string `json:"vcs_provider"`
}

type DeepSource struct {
	Host url.URL
}

type Runner struct {
	ID            string
	Host          url.URL
	ClientID      string
	ClientSecret  string
	WebhookSecret string
}

type Payload struct {
	RunnerID      string `json:"runner_id"`
	BaseURL       string `json:"base_url"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	WebhookSecret string `json:"webhook_secret"`
	Apps          []App  `json:"apps"`
}

type Syncer struct {
	client     *http.Client
	deepsource *DeepSource
	apps       []App
	runner     *Runner
	signer     Signer
}

func New(deepsource *DeepSource, runner *Runner, apps []App, signer Signer, client *http.Client) *Syncer {
	if client == nil {
		client = http.DefaultClient
	}

	return &Syncer{client: client, runner: runner, deepsource: deepsource, apps: apps, signer: signer}
}

func (s *Syncer) Sync() error {
	payload := &Payload{
		RunnerID:      s.runner.ID,
		BaseURL:       s.runner.Host.String(),
		ClientID:      s.runner.ClientID,
		ClientSecret:  s.runner.ClientSecret,
		WebhookSecret: s.runner.WebhookSecret,
		Apps:          s.apps,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to sync to DeepSource: %w", err)
	}

	target := s.deepsource.Host.JoinPath("/api/runner/").String()
	request, err := http.NewRequest(http.MethodPut, target, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to sync to DeepSource: %w", err)
	}

	token, err := s.signer.GenerateToken(s.runner.ID, []string{"sync"}, nil, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to sync to DeepSource: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Runner-ID", s.runner.ID)

	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to sync to DeepSource: %w", err)
	}

	if !(response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated) {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to sync to DeepSource: %w", err)
		}
		return fmt.Errorf("failed to sync to DeepSource: code=%d, body=%s", response.StatusCode, string(body))
	}

	return response.Body.Close()

}
