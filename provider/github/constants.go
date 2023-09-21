package github

import "errors"

const (
	HeaderGithubSignature = "x-hub-signature-256"
	HeaderRunnerSignature = "x-deepsource-signature-256"
	HeaderRunnerID        = "x-deepsource-runner-id"
	HeaderAppID           = "x-deepsource-app-id"
	HeaderContentType     = "Content-Type"
	HeaderInstallationID  = "X-Installation-Id"
)

var (
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrMandatoryArgsMissing = errors.New("mandatory args missing")
)
