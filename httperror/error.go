package httperror

import "net/http"

type Error struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func New(code int, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

var (
	ErrUnknown = func(err error) *Error {
		return New(http.StatusInternalServerError, "unknown error", err)
	}

	ErrMissingParams = func(err error) *Error {
		return New(http.StatusBadRequest, "mandatory parameters are missing or invalid", err)
	}

	ErrAppInvalid = func(err error) *Error {
		return New(http.StatusNotFound, "invalid app or credentials", err)
	}

	ErrAppUnsupported = func(err error) *Error {
		return New(http.StatusBadRequest, "app is not supported", err)
	}

	ErrUnauthorized = func(err error) *Error {
		return New(http.StatusUnauthorized, "unauthorized", err)
	}

	ErrBadRequest = func(err error) *Error {
		return New(http.StatusBadRequest, "bad request", err)
	}

	ErrUpstreamFailed = func(err error) *Error {
		return New(http.StatusBadGateway, "failed to proxy request", err)
	}
)
