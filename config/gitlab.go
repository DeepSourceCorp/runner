package config

import (
	"fmt"
	"net/url"
)

type Gitlab struct {
	AppID  string
	Secret string
	Host   url.URL
}

func (g *Gitlab) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type T struct {
		AppID   *string `yaml:"appId"`
		Secret  *string `yaml:"secret"`
		HostStr string  `yaml:"host"`
	}

	var v T
	if err := unmarshal(&v); err != nil {
		return fmt.Errorf("error unmarshalling gitlab config: %w", err)
	}

	host, err := url.Parse(v.HostStr)
	if err != nil {
		return fmt.Errorf("error unmarshalling gitlab config: %w", err)
	}

	g.AppID = *v.AppID
	g.Secret = *v.Secret
	g.Host = *host

	return nil
}
