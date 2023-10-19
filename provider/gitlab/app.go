package gitlab

import (
	"fmt"
	"net/url"
	"strings"
)

type App struct {
	ID            string
	WebhookSecret string
	APIHost       *url.URL
}

func (a *App) StripAPIURL(path string) *url.URL {
	prefix := fmt.Sprintf("/apps/%s/api/", a.ID)
	return a.APIHost.JoinPath(strings.TrimPrefix(path, prefix))
}

func (a *App) ValidateWebhookToken(token string) error {
	if token != a.WebhookSecret {
		return fmt.Errorf("webhook token mismatch")
	}
	return nil
}
