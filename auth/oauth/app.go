package oauth

import (
	"net/url"
)

type App struct {
	ID           string
	ClientID     string
	ClientSecret string
	RedirectURL  url.URL
	AuthHost     url.URL
	APIHost      url.URL
	Provider     string
}
