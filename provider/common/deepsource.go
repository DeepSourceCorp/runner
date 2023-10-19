package common

import "net/url"

type DeepSource struct {
	Host url.URL
}

func (d *DeepSource) WebhookURL(provider string) *url.URL {
	switch provider {
	case "github":
		return d.Host.JoinPath("/services/webhooks/github/")
	case "gitlab":
		return d.Host.JoinPath("/services/webhooks/gitlab/")
	case "bitbucket":
		return d.Host.JoinPath("/services/webhooks/bitbucket/")
	default:
		return nil
	}
}
