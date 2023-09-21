package github

import (
	"fmt"
	"net/http"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/proxyutil"
)

type APIService struct {
	appFactory *AppFactory
	client     *http.Client
}

func NewAPIService(appFactory *AppFactory, client *http.Client) *APIService {
	return &APIService{
		appFactory: appFactory,
		client:     client,
	}
}

func (s *APIService) Process(req *APIRequest) (*http.Response, error) {
	app := s.appFactory.GetApp(req.AppID)
	if app == nil {
		return nil, httperror.ErrAppInvalid(nil)
	}

	installationClient := NewInstallationClient(app, req.InstallationID, s.client)

	accessToken, err := installationClient.AccessToken()
	if err != nil {
		err := fmt.Errorf("failed to generate access token: %w", err)
		return nil, httperror.ErrUnknown(err)
	}

	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	forwarder := proxyutil.NewForwarder(s.client)
	res, err := forwarder.Forward(
		req.HTTPRequest,
		&proxyutil.ForwarderOpts{
			TargetURL: *installationClient.ProxyURL(req.HTTPRequest.URL.Path),
			Headers:   header,
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to proxy request: %w", err)
		return nil, httperror.ErrUnknown(err)
	}
	return res, nil
}