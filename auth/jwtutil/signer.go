package jwtutil

import (
	"crypto/rsa"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type Signer struct {
	Issuer     string
	privateKey *rsa.PrivateKey
}

func NewSigner(issuer string, privateKey *rsa.PrivateKey) *Signer {
	return &Signer{
		Issuer:     issuer,
		privateKey: privateKey,
	}
}

func (s *Signer) GenerateToken(scope []string, claims map[string]interface{}, expiry time.Duration) (string, error) {
	c := jwt.MapClaims{
		"iat": jwt.TimeFunc().Unix(),
		"exp": jwt.TimeFunc().Add(expiry).Unix(),
		"iss": s.Issuer,
		"scp": strings.Join(scope, " "),
	}
	for k, v := range claims {
		c[k] = v
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, c).SignedString(s.privateKey)
}
