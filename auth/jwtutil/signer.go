package jwtutil

import (
	"crypto/rsa"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type Signer struct {
	privateKey *rsa.PrivateKey
}

func NewSigner(privateKey *rsa.PrivateKey) *Signer {
	return &Signer{
		privateKey: privateKey,
	}
}

func (s *Signer) GenerateToken(issuer string, scope []string, claims map[string]interface{}, expiry time.Duration) (string, error) {
	c := jwt.MapClaims{
		"iat": jwt.TimeFunc().Unix(),
		"exp": jwt.TimeFunc().Add(expiry).Unix(),
		"scp": strings.Join(scope, " "),
	}

	if issuer != "" {
		c["iss"] = issuer
	}

	for k, v := range claims {
		c[k] = v
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, c).SignedString(s.privateKey)
}
