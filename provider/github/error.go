package github

import (
	"errors"
	"net/http"
)

var (
	ErrAppNotFound    = errors.New("could not find app")
	ErrInvalidHandler = errors.New("invalid handler")
)

type HTTPError struct {
	internal error  `json:"-"`
	Message  string `json:"error"`
	Code     int    `json:"-"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func ErrEchoResponse(err error) (int, interface{}) {
	e, ok := err.(*HTTPError)
	if !ok {
		return http.StatusInternalServerError, HTTPErrUnknown.WithInternal(err)
	}
	return e.Code, e
}

func (he *HTTPError) WithInternal(err error) *HTTPError {
	he.internal = err
	return he
}

var (
	HTTPErrUnknown           = &HTTPError{Code: http.StatusInternalServerError, Message: "unknown error"}
	HTTPErrAppNotFound       = &HTTPError{Code: http.StatusNotFound, Message: "invalid webhook url or the app does not exist"}
	HTTPErrInvalidRequest    = &HTTPError{Code: http.StatusBadRequest, Message: "invalid request"}
	HTTPErrSignatureMismatch = &HTTPError{Code: http.StatusUnauthorized, Message: "signature mismatch"}
	HTTPErrUpstreamFailed    = &HTTPError{Code: http.StatusBadGateway, Message: "failed to proxy request"}
	HTTPErrUpstreamBad       = &HTTPError{Code: http.StatusBadGateway, Message: "upstream returned bad response"}
)
