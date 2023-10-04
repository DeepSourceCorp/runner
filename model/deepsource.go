package model

import (
	"crypto/rsa"
	"net/url"

	"github.com/deepsourcecorp/runner/jwtutil"
)

type DeepSource struct {
	Host      url.URL
	PublicKey *rsa.PublicKey
}

func (d *DeepSource) Verifier() *jwtutil.Verifier {
	return jwtutil.NewVerifier(d.PublicKey)
}
