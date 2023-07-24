package config

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
		v.Namespace = "default"
	}

	k.Namespace = v.Namespace
	k.NodeSelector = v.NodeSelector
	k.ImageRegistry = v.ImageRegistry
	return nil
}
