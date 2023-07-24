package auth

import (
	"crypto/rsa"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type Signer struct {
	RunnerID  string
	SignerKey *rsa.PrivateKey
}

func NewSigner(runnerID string, signerKey *rsa.PrivateKey) *Signer {
	return &Signer{
		RunnerID:  runnerID,
		SignerKey: signerKey,
	}
}

func (s *Signer) GenerateToken(scope []string, expiry time.Duration) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": jwt.TimeFunc().Unix(),
		"exp": jwt.TimeFunc().Add(expiry).Unix(),
		"iss": s.RunnerID,
		"scp": strings.Join(scope, " "),
	}).SignedString(s.SignerKey)
}
