package config

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Runner        *Runner        `yaml:"runner"`
	DeepSource    *DeepSource    `yaml:"deepsource"`
	Apps          []*App         `yaml:"apps"`
	Kubernetes    *Kubernetes    `yaml:"kubernetes"`
	RQLite        *RQLite        `yaml:"rqlite"`
	SAML          *SAML          `yaml:"saml"`
	ObjectStorage *ObjectStorage `yaml:"objectStorage"`
	Sentry        *Sentry        `yaml:"sentry"`
}

func LoadConfig(r io.Reader) (*Config, error) {
	c := new(Config)
	if err := yaml.NewDecoder(r).Decode(c); err != nil {
		return nil, fmt.Errorf("config: failed to load config: %w", err)
	}
	if c.Kubernetes == nil {
		c.Kubernetes = &Kubernetes{}
		if err := c.Kubernetes.ParseFromEnv(); err != nil {
			return nil, err
		}
	}

	if c.RQLite == nil {
		c.RQLite = &RQLite{}
		if err := c.RQLite.ParseFromEnv(); err != nil {
			return nil, fmt.Errorf("config: failed to load config: %w", err)
		}
	}

	return c, nil
}
