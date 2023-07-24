package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"golang.org/x/exp/slog"
)

var (
	ErrSigningFailed = errors.New("signing failed")
	ErrMissingSecret = errors.New("secret cannot be empty")
)

// Signer is an interface for signing and verifying a payload.
type Signer interface {
	Sign(payload []byte) (string, error)
	Verify(payload []byte, signature string) error
}

type SHA256Signer struct {
	secret []byte
}

func NewSHA256Signer(secret []byte) (Signer, error) {
	if len(secret) == 0 {
		slog.Error("attempting to create signer with empty secret")
		return nil, ErrMissingSecret
	}
	return &SHA256Signer{
		secret: secret,
	}, nil
}

func (s *SHA256Signer) Sign(payload []byte) (string, error) {
	// generate HMAC hexdigest of payload.
	mac := hmac.New(sha256.New, s.secret)

	_, err := mac.Write(payload)
	if err != nil {
		return "", ErrSigningFailed
	}

	return "sha256=" + hex.EncodeToString(mac.Sum(nil)), nil
}

func (s *SHA256Signer) Verify(payload []byte, signature string) error {
	// compare HMAC hexdigest of payload.
	sig, err := s.Sign(payload)
	if err != nil {
		return err
	}

	if signature != sig {
		return errors.New("signature mismatch")
	}
	return nil
}
