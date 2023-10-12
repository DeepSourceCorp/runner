package common

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	runner := &Runner{
		ID:         "runner-id",
		PrivateKey: privateKey,
	}

	t.Run("issue and parse token", func(t *testing.T) {
		token, err := runner.IssueToken("code", map[string]interface{}{}, time.Hour)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		claims, err := runner.ParseToken(token)
		require.NoError(t, err)

		assert.Equal(t, "runner-id", claims["iss"])
	})
}
