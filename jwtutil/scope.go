package jwtutil

import (
	"strings"
)

const (
	ClaimScope = "scp"
)

func IsValidScope(claims map[string]interface{}, expectedScope string) bool {
	scope, ok := claims[ClaimScope].(string)
	if !ok {
		return false
	}

	return strings.Contains(scope, expectedScope)
}
