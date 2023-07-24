package github

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/deepsourcecorp/runner/internal/signer"
	"golang.org/x/exp/slog"
)

const (
	HeaderGithubSignature = "x-hub-signature-256"
	HeaderRunnerSignature = "x-deepsource-signature-256"
	HeaderRunnerID        = "x-deepsource-runner-id"
	HeaderAppID           = "x-deepsource-app-id"
	HeaderContentType     = "Content-Type"
)

var (
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrMandatoryArgsMissing = errors.New("mandatory args missing")
)

type WebhookProxyFactory struct {
	apps       map[string]*App
	deepsource *DeepSource
	runner     *Runner
	client     *http.Client
}

func NewWebhookProxyFactory(runner *Runner, deepsource *DeepSource, apps map[string]*App, client *http.Client) *WebhookProxyFactory {
	return &WebhookProxyFactory{
		runner:     runner,
		deepsource: deepsource,
		apps:       apps,
		client:     client,
	}
}

func (g *WebhookProxyFactory) New(appID string) (*WebhookProxy, error) {
	app := g.apps[appID]
	if app == nil {
		return nil, ErrAppNotFound
	}
	return NewWebhookProxy(app, g.runner, g.deepsource, g.client)
}

type WebhookProxy struct {
	app        *App
	runner     *Runner
	deepsource *DeepSource
	client     *http.Client
}

func NewWebhookProxy(app *App, runner *Runner, deepsource *DeepSource, client *http.Client) (*WebhookProxy, error) {
	return &WebhookProxy{
		app:        app,
		runner:     runner,
		deepsource: deepsource,
		client:     client,
	}, nil
}

func (p *WebhookProxy) VerifySignature(signature string, body []byte) error {
	signer, err := signer.NewSHA256Signer([]byte(p.app.WebhookSecret))
	if err != nil {
		slog.Error("failed to create signer", err)
		return err
	}
	if err := signer.Verify(body, signature); err != nil {
		return ErrInvalidSignature
	}
	return nil
}

func (p *WebhookProxy) webhookSignature(body []byte) (string, error) {
	signer, err := signer.NewSHA256Signer([]byte(p.runner.WebhookSecret))
	if err != nil {
		slog.Error("failed to create signer", err)
		return "", err
	}
	return signer.Sign(body)
}

func (p *WebhookProxy) Proxy(r *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	signature, err := p.webhookSignature(body)
	if err != nil {
		return nil, err
	}

	upstream := p.deepsource.Host.JoinPath("/services/webhooks/github/").String()
	req, err := http.NewRequest(
		r.Method,
		upstream,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	slog.Info("proxying request", "method", r.Method, "url", upstream)

	// copy query params
	q := req.URL.Query()
	for k, v := range r.URL.Query() {
		q[k] = v
	}
	req.URL.RawQuery = q.Encode()

	// copy headers
	for k, v := range r.Header {
		if strings.HasPrefix(k, "X-") {
			req.Header.Set(k, v[0])
		}
	}

	req.Header.Set(HeaderRunnerID, p.runner.ID)
	req.Header.Set(HeaderAppID, p.app.ID)
	req.Header.Set(HeaderRunnerSignature, signature)
	req.Header.Set(HeaderContentType, "application/json")

	res, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, err
}
