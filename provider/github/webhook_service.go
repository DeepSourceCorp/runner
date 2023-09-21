package github

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/provider/model"
	"github.com/deepsourcecorp/runner/proxyutil"
)

type WebhookService struct {
	appFactory *AppFactory
	runner     *model.Runner
	deepsource *model.DeepSource
	client     *http.Client
}

func NewWebhookService(appFactory *AppFactory, runner *model.Runner, deepsource *model.DeepSource, client *http.Client) *WebhookService {
	return &WebhookService{
		appFactory: appFactory,
		runner:     runner,
		deepsource: deepsource,
		client:     client,
	}
}

// Process processes the webhook request.  It verifies the signature, adds a
// signature for the cloud server, and then proxies the request to the cloud.
func (s *WebhookService) Process(req *WebhookRequest) (*http.Response, error) {
	app := s.appFactory.GetApp(req.AppID)
	if app == nil {
		return nil, httperror.ErrAppInvalid(nil)
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

	header := s.prepareHeader(app, signature)

	forwarder := proxyutil.NewForwarder(s.client)

	res, err := forwarder.Forward(req.HTTPRequest, &proxyutil.ForwarderOpts{
		TargetURL: *s.deepsource.WebhookURL(),
		Headers:   header,
		Query:     nil,
	})

	if err != nil {
		err := fmt.Errorf("failed to proxy request: %w", err)
		return nil, httperror.ErrUpstreamFailed(err)
	}

	return res, nil
}

func (s *WebhookService) prepareHeader(app *App, signature string) http.Header {
	header := http.Header{}
	header.Set(HeaderRunnerID, s.runner.ID)
	header.Set(HeaderAppID, app.ID)
	header.Set(HeaderRunnerSignature, signature)
	header.Set(HeaderContentType, "application/json")
	return header
}
