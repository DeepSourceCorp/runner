package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Kubernetes struct {
	Namespace     string            `yaml:"namespace"`
	NodeSelector  map[string]string `yaml:"nodeSelector"`
	ImageRegistry *ImageRegistry    `yaml:"imageRegistry"`
}

func (k *Kubernetes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type T struct {
		Namespace     string            `yaml:"namespace"`
		NodeSelector  map[string]string `yaml:"nodeSelector"`
		ImageRegistry *ImageRegistry    `yaml:"imageRegistry"`
	}
	var v T
	if err := unmarshal(&v); err != nil {
		return err
	}

	if v.Namespace == "" {
		v.Namespace = os.Getenv("TASK_NAMESPACE")
	}

	if v.NodeSelector == nil {
		ns := os.Getenv("TASK_NODE_SELECTOR")
		v.NodeSelector = make(map[string]string)
		err := yaml.Unmarshal([]byte(ns), &v.NodeSelector)
		if err != nil {
			return err
		}
	}

	if v.ImageRegistry == nil {
		imageRegistry := &ImageRegistry{}
		err := imageRegistry.ParseFromEnv()
		if err != nil {
			return err
		}
		v.ImageRegistry = imageRegistry
	}

	k.Namespace = v.Namespace
	k.NodeSelector = v.NodeSelector
	k.ImageRegistry = v.ImageRegistry
	return nil
}

func (k *Kubernetes) ParseFromEnv() error {
	if os.Getenv("TASK_NAMESPACE") != "" {
		k.Namespace = os.Getenv("TASK_NAMESPACE")
	}
	if os.Getenv("TASK_NODE_SELECTOR") != "" {
		ns := os.Getenv("TASK_NODE_SELECTOR")
		k.NodeSelector = make(map[string]string)
		err := yaml.Unmarshal([]byte(ns), &k.NodeSelector)
		if err != nil {
			return err
		}
	}

	imageRegistry := &ImageRegistry{}
	err := imageRegistry.ParseFromEnv()
	if err != nil {
		return err
	}
	k.ImageRegistry = imageRegistry

	return nil
}
