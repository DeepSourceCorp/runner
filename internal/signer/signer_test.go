package signer

import (
	"testing"
)

func TestSHA256Signer_Sign(t *testing.T) {
	s := &SHA256Signer{
		secret: []byte("test-secret"),
	}

	signature, err := s.Sign([]byte("test-payload"))
	if err != nil {
		t.Errorf("SHA256Signer.Sign() error = %v", err)
		return
	}

	if signature != "sha256=5b12467d7c448555779e70d76204105c67d27d1c991f3080c19732f9ac1988ef" {
		t.Errorf("SHA256Signer.Sign() = %v, want %v", signature, "sha256=5b12467d7c448555779e70d76204105c67d27d1c991f3080c19732f9ac1988ef")
	}
}

func TestSHA256Signer_Verify(t *testing.T) {
	s := &SHA256Signer{
		secret: []byte("test-secret"),
	}

	err := s.Verify([]byte("test-payload"), "sha256=5b12467d7c448555779e70d76204105c67d27d1c991f3080c19732f9ac1988ef")
	if err != nil {
		t.Errorf("SHA256Signer.Verify() unexpected error, got = %v", err)
		return
	}

	err = s.Verify([]byte("test-payload"), "invalid-signature")
	if err == nil {
		t.Errorf("SHA256Signer.Verify(), expected error, got = %v", err)
		return
	}
}
