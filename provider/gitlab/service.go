package gitlab

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/deepsourcecorp/runner/forwarder"
	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/provider/common"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

type TokenResolver interface {
	GetToken(token string) (*oauth2.Token, error)
}

type Service struct {
	runner        *common.Runner
	deepsource    *common.DeepSource
	apps          map[string]*App
	tokenResolver TokenResolver
	client        *http.Client
}

type ServiceOpts struct {
	Runner        *common.Runner
	DeepSource    *common.DeepSource
	Apps          map[string]*App
	TokenResolver TokenResolver
	Client        *http.Client
}

func NewService(opts *ServiceOpts) *Service {
	if opts.Client == nil {
		opts.Client = http.DefaultClient
	}
	return &Service{
		runner:        opts.Runner,
		deepsource:    opts.DeepSource,
		apps:          opts.Apps,
		client:        opts.Client,
		tokenResolver: opts.TokenResolver,
	}
}

func (s *Service) ForwardAPI(req *APIRequest) (*http.Response, error) {
	app := s.apps[req.AppID]
	if app == nil {
		slog.Error("[gitlab.Service] app not found", slog.String("app_id", req.AppID))
		return nil, httperror.ErrBadRequest(nil)
	}

	token, err := s.tokenResolver.GetToken(req.Token)
	if err != nil {
		slog.Error("[gitlab.Service] failed to resolve token", slog.Any("err", err))
		return nil, err
	}

	// Remove the DeepSource authorization header if present
	req.HTTPRequest.Header.Del(common.HeaderAuthorization)

	header := http.Header{}
	header.Set(common.HeaderAuthorization, fmt.Sprintf("Bearer %s", token.AccessToken))

	res, err := forwarder.Forward(
		req.HTTPRequest,
		&forwarder.Opts{
			TargetURL: app.StripAPIURL(req.HTTPRequest.URL.Path),
			Headers:   header,
		}, s.client)

	slog.Info("Status code from Gitlab", slog.Int("status_code", res.StatusCode))

	if err != nil {
		slog.Error("[APIService.Process] failed to proxy request", slog.Any("err", err))
		return nil, httperror.ErrUnknown(err)
	}
	return res, nil
}

func (s *Service) ForwardWebhook(req *WebhookRequest) (*http.Response, error) {
	app := s.apps[req.AppID]
	if app == nil {
		slog.Error("[gitlab.Service] app not found", slog.String("app_id", req.AppID))
		return nil, httperror.ErrBadRequest(nil)
	}

	if err := app.ValidateWebhookToken(req.Token); err != nil {
		slog.Error("[WebhookService.Process] failed to validate webhook token", slog.Any("err", err))
		return nil, httperror.ErrUnauthorized(err)
	}

	body, err := io.ReadAll(req.HTTPRequest.Body)
	if err != nil {
		return nil, httperror.ErrUnknown(
			fmt.Errorf("failed to read request body: %w", err),
		)
	}
	req.HTTPRequest.Body = io.NopCloser(bytes.NewReader(body)) // rewind body

	signature, err := s.runner.SignPayload(body)
	if err != nil {
		slog.Error("[WebhookService.Process] failed to sign payload", slog.Any("err", err))
		return nil, httperror.ErrUnknown(err)
	}

	extraHeaders := http.Header{}
	extraHeaders.Set(common.HeaderRunnerSignature, signature)
	extraHeaders.Set(common.HeaderAppID, req.AppID)
	extraHeaders.Set(common.HeaderRunnerID, s.runner.ID)

	res, err := forwarder.Forward(req.HTTPRequest, &forwarder.Opts{
		TargetURL: s.deepsource.WebhookURL(common.ProviderGitlab),
		Headers:   extraHeaders,
	}, s.client)

	slog.Info("Status code from Gitlab", slog.Int("status_code", res.StatusCode))

	if err != nil {
		slog.Error("[WebhookService.Process] failed to proxy request", slog.Any("err", err))
		return nil, httperror.ErrUnknown(err)
	}
	return res, nil
}

func (s *Service) RemoteURL(req interface{}) (string, error) {
	r, ok := req.(*RemoteURLRequest)
	if !ok {
		slog.Error("[gitlab.Service] invalid request type for RemoteURL()", slog.Any("req", req))
		return "", errors.New("invalid request type for ReniteURL")
	}
	token, err := s.tokenResolver.GetToken(r.Token)
	if err != nil {
		slog.Error("[gitlab.Service] failed to resolve token", slog.Any("err", err))
		return "", err
	}

	u, err := url.Parse(r.SourceURL)
	if err != nil {
		slog.Error("[gitlab.Service] failed to parse source url", slog.Any("err", err))
		return "", err
	}

	u.User = url.UserPassword("oauth2", token.AccessToken)
	return u.String(), nil
}
