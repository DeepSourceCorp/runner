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
	if os.Getenv("TASK_IMAGE_PULL_SECRET_NAME") != "" {
		i.PullSecretName = os.Getenv("TASK_IMAGE_PULL_SECRET_NAME")
	}

	if os.Getenv("TASK_IMAGE_REGISTRY_URL") != "" {
		imageURL, err := url.Parse(os.Getenv("TASK_IMAGE_REGISTRY_URL"))
		if err != nil {
			return err
		}
		i.RegistryUrl = *imageURL
	}
	return nil
}
