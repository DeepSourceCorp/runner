package saml

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"

	"github.com/crewjam/saml/samlsp"
	"golang.org/x/exp/slog"
)

type Opts struct {
	Certificate *tls.Certificate
	MetadataURL url.URL
	RootURL     url.URL
}

func NewSAMLMiddleware(ctx context.Context, opts *Opts, client *http.Client) (*samlsp.Middleware, error) {
	idpMetadata, err := samlsp.FetchMetadata(ctx, client, opts.MetadataURL)
	if err != nil {
		return nil, err
	}

	privateKey, ok := opts.Certificate.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		slog.Info("private key is not an RSA key")
		return nil, errors.New("private key is not an RSA key")
	}

	sp, err := samlsp.New(samlsp.Options{
		URL:               opts.RootURL,
		Key:               privateKey,
		Certificate:       opts.Certificate.Leaf,
		IDPMetadata:       idpMetadata,
		AllowIDPInitiated: true,
	})

	return sp, err
}
