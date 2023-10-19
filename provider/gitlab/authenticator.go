package gitlab

type Authenticator struct {
	service *Service
}

func NewAuthenticator(service *Service) *Authenticator {
	return &Authenticator{
		service: service,
	}
}

func (a *Authenticator) RemoteURL(_, sourceURL string, extra map[string]interface{}) (string, error) {
	req := NewRemoteURLRequest(sourceURL, extra)
	return a.service.RemoteURL(req)
}
