package config

type ObjectStorage struct {
	Provider   string `yaml:"provider"`
	Bucket     string `yaml:"bucket"`
	Credential string `yaml:"credential"`
}
