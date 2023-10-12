package common

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/jwtutil"
)

type Runner struct {
	ID           string
	ClientID     string
	ClientSecret string
	PrivateKey   *rsa.PrivateKey
}

func (r *Runner) ValidateClientID(clientID string) error {
	if r.ClientID == "" || r.ClientID != clientID {
		err := fmt.Errorf("client id is empty")
		return httperror.ErrBadRequest(err)
	}
	return nil
}

func (r *Runner) ValidateClientSecret(clientSecret string) error {
	if r.ClientSecret == "" || r.ClientSecret != clientSecret {
		err := fmt.Errorf("client secret is empty")
		return httperror.ErrBadRequest(err)
	}
	return nil
}

func (r *Runner) ValidateClient(clientID, clientSecret string) error {
	if err := r.ValidateClientID(clientID); err != nil {
		return err
	}
	if err := r.ValidateClientSecret(clientSecret); err != nil {
		return err
	}
	return nil
}

func (r *Runner) IssueToken(scope string, claims map[string]interface{}, expiry time.Duration) (string, error) {
	signer := jwtutil.NewSigner(r.PrivateKey)
	token, err := signer.GenerateToken(
		r.ID, []string{scope}, claims, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate jwt token, %w", err)
	}
	return token, nil
}

func (r *Runner) ParseToken(token string) (map[string]interface{}, error) {
	verifier := jwtutil.NewVerifier(&r.PrivateKey.PublicKey)
	claims, err := verifier.Verify(token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify jwt token, %w", err)
	}
	return claims, nil
}
