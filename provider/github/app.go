package github

import (
	"crypto/rsa"
	"fmt"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v4"
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

func (a *App) JWT() (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": jwt.TimeFunc().Add(-1 * time.Minute).Unix(),
		"exp": jwt.TimeFunc().Add(10 * time.Minute).Unix(),
		"iss": a.AppID,
	}).SignedString(a.PrivateKey)
}

func (a *App) InstallationURL() string {
	return a.BaseHost.JoinPath(fmt.Sprintf("/apps/%s/installations/new", a.AppSlug)).String()
}
