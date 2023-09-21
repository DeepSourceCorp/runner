package github

import (
	"crypto/rsa"
	"fmt"
	"net/url"
	"time"

	"github.com/deepsourcecorp/runner/internal/signer"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/exp/slog"
)

type App struct {
	ID string

	// Github App configuration.
	AppID         string
	AppSlug       string
	WebhookSecret string
	BaseHost      url.URL
	APIHost       url.URL
	PrivateKey    *rsa.PrivateKey
}

// Generate a JWT token for the GitHub App.
// (https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-json-web-token-jwt-for-a-github-app)
func (a *App) JWT() (string, error) {
	if a.PrivateKey == nil {
		return "", fmt.Errorf("private key is required")
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": jwt.TimeFunc().Add(-1 * time.Minute).Unix(),
		"exp": jwt.TimeFunc().Add(10 * time.Minute).Unix(),
		"iss": a.AppID,
	}).SignedString(a.PrivateKey)
}

// InstallationURL returns the URL to install the GitHub App.
func (a *App) InstallationURL() string {
	return a.BaseHost.JoinPath(fmt.Sprintf("/apps/%s/installations/new", a.AppSlug)).String()
}

// VerifyWebhookSignature verifies the signature of the webhook payload.
// (https://docs.github.com/en/developers/webhooks-and-events/securing-your-webhooks#validating-payloads-from-github)
func (a *App) VerifyWebhookSignature(payload []byte, signature string) error {
	signer, err := signer.NewSHA256Signer([]byte(a.WebhookSecret))
	if err != nil {
		slog.Error("failed to create signer", err)
		return err
	}
	if err := signer.Verify(payload, signature); err != nil {
		return ErrInvalidSignature
	}
	return nil
}
