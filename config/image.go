package config

import (
	"net/url"
	"os"
)

type ImageRegistry struct {
	PullSecretName string  `yaml:"pullSecretName"`
	RegistryUrl    url.URL `yaml:"registryUrl"`
}

func (i *ImageRegistry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type T struct {
		PullSecretName string `yaml:"pullSecretName"`
		RegistryUrl    string `yaml:"registryUrl"`
	}
	var v T
	if err := unmarshal(&v); err != nil {
		return err
	}
	if os.Getenv("TASK_IMAGE_PULL_SECRET_NAME") != "" {
		i.PullSecretName = os.Getenv("TASK_IMAGE_PULL_SECRET_NAME")
	} else {
		i.PullSecretName = v.PullSecretName
	}

	imageURL, err := url.Parse(v.RegistryUrl)
	if err != nil {
		return err
	}

	i.RegistryUrl = *imageURL
	return nil
}
