package common

import "time"

const (
	ExpiryRunnerAccessToken  = time.Minute * 5
	ExpiryRunnerRefreshToken = time.Hour * 24 * 30

	CookieNameSession = "session"
	CookieNameRefresh = "refresh"
)
