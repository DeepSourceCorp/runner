package github

type AppFactory struct {
	apps map[string]*App
}

func NewAppFactory(apps map[string]*App) *AppFactory {
	return &AppFactory{
		apps: apps,
	}
}

func (f *AppFactory) GetApp(appID string) *App {
	return f.apps[appID]
}
