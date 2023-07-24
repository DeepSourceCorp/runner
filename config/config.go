package config

import (
	"io"

	"golang.org/x/exp/slog"
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
}

func LoadConfig(r io.Reader) (*Config, error) {
	c := new(Config)
	if err := yaml.NewDecoder(r).Decode(c); err != nil {
		slog.Error("failed to unmarshal config file", slog.Any("err", err))
		return nil, err
	}
	return c, nil
}
