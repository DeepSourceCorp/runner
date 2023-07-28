package model

import (
	"crypto/rsa"
	"net/url"
)

type DeepSource struct {
	Host      url.URL
	PublicKey *rsa.PublicKey
}
