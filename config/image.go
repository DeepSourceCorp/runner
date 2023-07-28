package config

import (
	"errors"
	"net/url"
	"os"
)

type ImageRegistry struct {
	PullSecretName string  `yaml:"pullSecretName"`
	RegistryUrl    url.URL `yaml:"registryUrl"`
}

func (i *ImageRegistry) ParseFromEnv() error {
	if os.Getenv("TASK_IMAGE_PULL_SECRET_NAME") == "" || os.Getenv("TASK_IMAGE_REGISTRY_URL") == "" {
		return errors.New("config: failed to parse image registry from env")
	}

	i.PullSecretName = os.Getenv("TASK_IMAGE_PULL_SECRET_NAME")

	imageURL, err := url.Parse(os.Getenv("TASK_IMAGE_REGISTRY_URL"))
	if err != nil {
		return err
	}
	i.RegistryUrl = *imageURL

	return nil
}
