package model

import "net/url"

type DeepSource struct {
	Host url.URL
}

func (d *DeepSource) WebhookURL() *url.URL {
	return d.Host.JoinPath("/services/webhooks/github/")
}
