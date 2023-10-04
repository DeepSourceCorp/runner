package model

import (
	"crypto/rsa"

	"github.com/deepsourcecorp/runner/jwtutil"
)

type Runner struct {
	ID            string
	ClientID      string
	ClientSecret  string
	PrivateKey    *rsa.PrivateKey
	WebhookSecret string
}

func (r *Runner) IsValidClientID(clientID string) bool {
	if r.ClientID == "" {
		return false
	}
	return r.ClientID == clientID
}

func (r *Runner) IsValidClientSecret(clientSecret string) bool {
	if r.ClientSecret == "" {
		return false
	}
	return r.ClientSecret == clientSecret
}

func (r *Runner) IsValidClient(clientID, clientSecret string) bool {
	return r.IsValidClientID(clientID) && r.IsValidClientSecret(clientSecret)
}

func (r *Runner) Signer() *jwtutil.Signer {
	return jwtutil.NewSigner(r.PrivateKey)
}

func (r *Runner) Verifier() *jwtutil.Verifier {
	return jwtutil.NewVerifier(&r.PrivateKey.PublicKey)
}
