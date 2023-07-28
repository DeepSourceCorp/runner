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
	k.Namespace = v.Namespace
	if os.Getenv("TASK_NAMESPACE") != "" {
		k.Namespace = os.Getenv("TASK_NAMESPACE")
	}
	k.NodeSelector = v.NodeSelector
	if os.Getenv("TASK_NODE_SELECTOR") != "" {
		ns := os.Getenv("TASK_NODE_SELECTOR")
		k.NodeSelector = make(map[string]string)
		err := yaml.Unmarshal([]byte(ns), &k.NodeSelector)
		if err != nil {
			return err
		}
	}
	k.ImageRegistry = v.ImageRegistry
	return nil
}
