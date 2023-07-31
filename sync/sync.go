package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

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
	client  *http.Client
	payload *Payload
	target  string
}

func New(deepsource *DeepSource, runner *Runner, apps []App, client *http.Client) *Syncer {
	if client == nil {
		client = http.DefaultClient
	}
	payload := &Payload{
		RunnerID:      runner.ID,
		BaseURL:       runner.Host.String(),
		ClientID:      runner.ClientID,
		ClientSecret:  runner.ClientSecret,
		WebhookSecret: runner.WebhookSecret,
		Apps:          apps,
	}

	target := deepsource.Host.JoinPath("/api/runners/").String()
	return &Syncer{client: client, payload: payload, target: target}
}

func (s *Syncer) Sync() error {
	data, err := json.Marshal(s.payload)
	if err != nil {
		return fmt.Errorf("failed to sync to DeepSource: %w", err)
	}
	request, err := http.NewRequest(http.MethodPut, s.target, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to sync to DeepSource: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to sync to DeepSource: %w", err)
	}

	if !(response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated) {
		return fmt.Errorf("failed to sync to DeepSource: status=%d", response.StatusCode)
	}

	return response.Body.Close()

}
