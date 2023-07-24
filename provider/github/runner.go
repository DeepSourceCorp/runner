package github

import (
	"crypto/rsa"
	"net/url"
)

type Runner struct {
	ID            string
	WebhookSecret string
}

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

type DeepSource struct {
	Host url.URL
}
