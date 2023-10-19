package github

import (
	"errors"
	"fmt"
)

const (
	HeaderGithubSignature   = "x-hub-signature-256"
	HeaderInstallationID    = "X-Installation-Id"
	HeaderValueGithubAccept = "application/vnd.github+json"
)

var (
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrMandatoryArgsMissing = errors.New("mandatory args missing")
	ErrAppNotFound          = fmt.Errorf("app not found")
)
