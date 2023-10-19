package github

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/deepsourcecorp/runner/forwarder"
	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/provider/common"
	"golang.org/x/exp/slog"
)

type Service struct {
	runner     *common.Runner
	deepsource *common.DeepSource
	apps       map[string]*App
	client     *http.Client
}

type ServiceOpts struct {
	Runner     *common.Runner
	DeepSource *common.DeepSource
	Apps       map[string]*App
	Client     *http.Client
}

func NewService(opts *ServiceOpts) *Service {
	if opts.Client == nil {
		opts.Client = http.DefaultClient
	}
	return &Service{
		runner:     opts.Runner,
		deepsource: opts.DeepSource,
		apps:       opts.Apps,
		client:     opts.Client,
	}
}

func (s *Service) ForwardAPI(req *APIRequest) (*http.Response, error) {
	app := s.apps[req.AppID]
	if app == nil {
		slog.Error("[github.Service] app not found", slog.String("app_id", req.AppID))
		return nil, httperror.ErrBadRequest(nil)
	}

	accessToken, err := GetAccessToken(app, req.InstallationID, s.client)
	if err != nil {
		slog.Error("[github.Service] failed to generate access token", slog.Any("err", err))
		return nil, httperror.ErrUnknown(err)
	}

	// Remove the DeepSource authorization header if present
	req.HTTPRequest.Header.Set("Authorization", "Bearer "+accessToken)

	extraHeaders := http.Header{}
	extraHeaders.Set(common.HeaderAuthorization, fmt.Sprintf("Bearer %s", accessToken))
	if req.HTTPRequest.Header.Get(common.HeaderAccept) == "" {
		extraHeaders.Set(common.HeaderAccept, HeaderValueGithubAccept)
	}

	res, err := forwarder.Forward(
		req.HTTPRequest,
		&forwarder.Opts{
			TargetURL: app.StripAPIURL(req.HTTPRequest.URL.Path),
			Headers:   extraHeaders,
		}, s.client)

	slog.Info("Status code from GitHub", slog.Int("status_code", res.StatusCode))

	if err != nil {
		err = fmt.Errorf("failed to proxy request: %w", err)
		return nil, httperror.ErrUnknown(err)
	}
	return res, nil
}

func (s *Service) ForwardWebhook(req *WebhookRequest) (*http.Response, error) {
	app := s.apps[req.AppID]
	if app == nil {
		slog.Error("[github.Service] app not found", slog.String("app_id", req.AppID))
		return nil, httperror.ErrBadRequest(nil)
	}

	// Read body and rewind.  We need the body to process signatures.  The body
	// is also needed to proxy the request to the cloud server.
	body, err := io.ReadAll(req.HTTPRequest.Body)
	if err != nil {
		return nil, httperror.ErrUnknown(
			fmt.Errorf("failed to read request body: %w", err),
		)
	}
	req.HTTPRequest.Body = io.NopCloser(bytes.NewReader(body)) // rewind body

	if err := app.VerifyWebhookSignature(body, req.Signature); err != nil {
		return nil, httperror.ErrUnauthorized(
			fmt.Errorf("failed to verify webhook signature: %w", err),
		)
	}

	// generate signature for cloud server
	signature, err := s.runner.SignPayload(body)
	if err != nil {
		err = fmt.Errorf("failed to sign payload: %w", err)
		return nil, httperror.ErrUnknown(err)
	}

	extraHeaders := http.Header{}
	extraHeaders.Set(common.HeaderRunnerID, s.runner.ID)
	extraHeaders.Set(common.HeaderAppID, app.ID)
	extraHeaders.Set(common.HeaderRunnerSignature, signature)
	extraHeaders.Set(common.HeaderContentType, "application/json")

	res, err := forwarder.Forward(req.HTTPRequest, &forwarder.Opts{
		TargetURL: s.deepsource.WebhookURL(common.ProviderGithub),
		Headers:   extraHeaders,
		Query:     nil,
	}, s.client)

	slog.Info("Status code from DeepSource", slog.Int("status_code", res.StatusCode))

	if err != nil {
		err := fmt.Errorf("failed to proxy request: %w", err)
		return nil, httperror.ErrUpstreamFailed(err)
	}

	return res, nil
}

func (s *Service) InstallationURL(req *InstallationRequest) (string, error) {
	app := s.apps[req.AppID]
	if app == nil {
		slog.Error("[github.Service] app not found", slog.String("app_id", req.AppID))
		return "", httperror.ErrBadRequest(nil)
	}

	return app.InstallationURL(), nil
}

// GetRemoteURL returns an a remote URL that is used to clone the repository.
// The URL is authenitcated with the installation token.
func (s *Service) RemoteURL(req *RemoteURLRequest) (string, error) {
	app := s.apps[req.AppID]
	if app == nil {
		slog.Error("[github.Service] app not found", slog.String("app_id", req.AppID))
		return "", httperror.ErrBadRequest(nil)
	}

	sourceURL, err := url.Parse(req.SourceURL)
	if err != nil {
		slog.Error("[github.Service] failed to parse source url", slog.Any("err", err))
		return "", fmt.Errorf("failed to parse url: %w", err)
	}

	token, err := GetAccessToken(app, req.InstallationID, s.client)
	if err != nil {
		slog.Error("[github.Service] failed to generate access token", slog.Any("err", err))
		return "", httperror.ErrUnknown(err)
	}

	sourceURL.User = url.UserPassword("x-access-token", token)
	return sourceURL.String(), nil
}
