package github

import (
	"crypto/rand"
	"crypto/rsa"
	"net/url"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestApp_JWT(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	app := &App{ID: "test-app-id", PrivateKey: privateKey}
	token, err := app.JWT()
	assert.NoError(t, err)

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		publicKey := privateKey.PublicKey
		return &publicKey, nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

}

func TestApp_InstallationURL(t *testing.T) {
	baseURL, _ := url.Parse("http://example.com")
	app := &App{BaseHost: *baseURL, AppSlug: "test-app-slug"}
	assert.Equal(t, "http://example.com/apps/test-app-slug/installations/new", app.InstallationURL())
}
