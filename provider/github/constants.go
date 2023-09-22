package github

import (
	"errors"
	"fmt"
)

const (
	HeaderGithubSignature = "x-hub-signature-256"
	HeaderRunnerSignature = "x-deepsource-signature-256"
	HeaderRunnerID        = "x-deepsource-runner-id"
	HeaderAppID           = "x-deepsource-app-id"
	HeaderInstallationID  = "X-Installation-Id"

	HeaderContentType    = "Content-Type"
	HeaderAuthorization  = "Authorization"
	HeaderAccept         = "Accept"
	HeaderAcceptEncoding = "Accept-Encoding"

	HeaderValueGithubAccept = "application/vnd.github+json"
)

var (
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrMandatoryArgsMissing = errors.New("mandatory args missing")
	ErrAppNotFound          = fmt.Errorf("app not found")
)
