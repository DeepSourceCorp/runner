package common

import "time"

type ContextKey string

const (
	ExpiryRunnerAccessToken  = time.Minute * 5
	ExpiryRunnerRefreshToken = time.Hour * 24 * 30

	CookieNameSession = "session"
	CookieNameRefresh = "refresh"

	ContextKeyRequestID ContextKey = "request_id"
)
