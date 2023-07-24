package config

import (
	"crypto/rsa"
	"net/url"

	"github.com/golang-jwt/jwt"
)

const (
	RunnerKeyFile = "runner.rsa"
)

type Runner struct {
	ID            string `json:"id"`
	ClientID      string `json:"clientId"`
	ClientSecret  string `json:"clientSecret"`
	WebhookSecret string `json:"webhookSecret"`
	Host          url.URL
	PrivateKey    *rsa.PrivateKey
}

func (r *Runner) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type T struct {
		ID            *string `yaml:"id"`
		ClientID      *string `yaml:"clientId"`
		ClientSecret  *string `yaml:"clientSecret"`
		WebhookSecret *string `yaml:"webhookSecret"`
		HostStr       string  `yaml:"host"`
		PrivateKeyStr string  `yaml:"privateKey"`
	}

	var v T
	if err := unmarshal(&v); err != nil {
		return err
	}
	h, err := url.Parse(v.HostStr)
	if err != nil {
		return err
	}
	k, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(v.PrivateKeyStr))
	if err != nil {
		return err
	}
	r.ID = *v.ID
	r.ClientID = *v.ClientID
	r.ClientSecret = *v.ClientSecret
	r.WebhookSecret = *v.WebhookSecret
	r.Host = *h
	r.PrivateKey = k
	return nil
}
