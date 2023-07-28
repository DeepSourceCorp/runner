package model

import (
	"net/url"
)

type Runner struct {
	ID            string
	WebhookSecret string
}

type DeepSource struct {
	Host url.URL
}
