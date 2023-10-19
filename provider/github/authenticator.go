package github

type Authenticator struct {
	service *Service
}

func NewAuthenticator(service *Service) *Authenticator {
	return &Authenticator{
		service: service,
	}
}

func (a *Authenticator) RemoteURL(appID, sourceURL string, extra map[string]interface{}) (string, error) {
	req := NewRemoteURLRequest(appID, sourceURL, extra)
	return a.service.RemoteURL(req)
}
