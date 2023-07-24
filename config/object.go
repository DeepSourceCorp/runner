package config

type ObjectStorage struct {
	Backend    string `yaml:"backend"`
	Bucket     string `yaml:"bucket"`
	Credential string `yaml:"credential"`
}
