package config

import (
	"crypto/rsa"
	"net/url"

	"github.com/golang-jwt/jwt"
)

type Github struct {
	AppID         string  `yaml:"appId"`
	ClientID      string  `yaml:"clientId"`
	ClientSecret  string  `yaml:"clientSecret"`
	WebhookSecret string  `yaml:"webhookSecret"`
	Host          url.URL `yaml:"-"`
	APIHost       url.URL `yaml:"-"`
	Slug          string  `yaml:"slug"`
	PrivateKey    *rsa.PrivateKey
}

func (g *Github) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type T struct {
		AppID         *string `yaml:"appId"`
		ClientID      *string `yaml:"clientId"`
		ClientSecret  *string `yaml:"clientSecret"`
		WebhookSecret *string `yaml:"webhookSecret"`
		HostStr       string  `yaml:"host"`
		APIHostStr    string  `yaml:"apiHost"`
		Slug          *string `yaml:"slug"`
		PrivateKeyStr string  `yaml:"privateKey"`
	}

	var v T
	if err := unmarshal(&v); err != nil {
		return err
	}
	host, err := url.Parse(v.HostStr)
	if err != nil {
		return err
	}
	apiHost, err := url.Parse(v.APIHostStr)
	if err != nil {
		return err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(v.PrivateKeyStr))
	if err != nil {
		return err
	}
	g.AppID = *v.AppID
	g.ClientID = *v.ClientID
	g.ClientSecret = *v.ClientSecret
	g.WebhookSecret = *v.WebhookSecret
	g.Host = *host
	g.APIHost = *apiHost
	g.Slug = *v.Slug
	g.PrivateKey = privateKey
	return nil
}
