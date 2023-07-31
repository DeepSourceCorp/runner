package config

const (
	AppConfigDir = "./apps"
)

type App struct {
	ID       string  `yaml:"id"`
	Name     string  `yaml:"name"`
	Provider string  `yaml:"provider"`
	Github   *Github `yaml:"github"`
}
