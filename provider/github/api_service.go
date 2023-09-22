package github

import (
	"fmt"
	"net/http"

	"github.com/deepsourcecorp/runner/forwarder"
	"github.com/deepsourcecorp/runner/httperror"
	"golang.org/x/exp/slog"
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

	// Remove the DeepSource authorization header if present
	req.HTTPRequest.Header.Del("Authorization")
	req.HTTPRequest.Header.Del("Accept-Encoding")

	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	if req.HTTPRequest.Header.Get("Accept") == "" {
		header.Set("Accept", HeaderValueGithubAccept)
	} else {
		header.Set("Accept", req.HTTPRequest.Header.Get("Accept"))
	}

	f := forwarder.New(s.client)
	res, err := f.Forward(
		req.HTTPRequest,
		&forwarder.Opts{
			TargetURL: *installationClient.ProxyURL(req.HTTPRequest.URL.Path),
			Headers:   header,
		})

	slog.Info("Status code from GitHub", slog.Int("status_code", res.StatusCode))

	if err != nil {
		err = fmt.Errorf("failed to proxy request: %w", err)
		return nil, httperror.ErrUnknown(err)
	}
	return res, nil
}
