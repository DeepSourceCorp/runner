package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRunner_UnmarshalYAML(t *testing.T) {
	input := `
id: "runner-id"
clientId: "client-id"
clientSecret: "client-secret"
webhookSecret: "webhook-secret"
host: "https://example.com"
privateKey: |
  -----BEGIN PRIVATE KEY-----
  MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQCr26gJM8hgt8i5
  SPTZg4Io+OioIJtQg7thJeDUMBLB32ycXB98INj6zFteX//PkuqD8lCENxbsOQDV
  CYUq0H+aky4TVFKXHif3p6gQTXkU/Yqp39EWm0DdMMerOI0r7J1zMsyYw/6jvSUW
  HAZYp4sEahWmn/T8ZC+mJZZ789FVunLyCbWtYl4DmEkkuA1olFacXf6hyKEjBZ3B
  Ow+dkJHfDUpCNmnb5j00Ev3Zb7mPCshoy2/NX5/e3+KZDZGDL9zbWkNsYoD1n910
  sVXGegPLU9TMUz5FQoSVPwaxf9j4sQTNPOp0Ilib8x8pWOfs7Iy3A3XXZNDlm9mQ
  jNhUvN+7AgMBAAECggEBAJRs1BaGg4OMlq33ZYhKPOrX9k/mQV1rODTx+tgnYLvS
  E8KDCaox0FPilPLQJGYIs8QbThCyZ3jCzoYvf7R3eA1vGbcV93KOV+RbBxp1XqKT
  SuPl6nYExiOCkp+86qfJ5j3s3Kj/dPfDTrlmoNCGetjoKiTLN1GX0VNEWVBaRiwr
  vJ0ptEPKYH60hoWr9yYb3ZC4Qp8uzy+eLa60SA+dDIGcsmsjEi/mINig5DOCLS4A
  xLpnmoR8n0ym4feJXWrgeRsxJDOu7fmsDstWCxNEdV685btsrMnXJsdfEG1i7jBU
  CgHJl+/6R3EXWwgNWO3sez5MiRjGd3/fpRlZUX3n31ECgYEA5J3FZJh9JYjWkO/M
  2jU5G8XUoHlmBOwVc9od5/fqESb9tTi1ZWGN95lBld7flMFkzqBUMAwitVK9XbBU
  3Kj4P/V3bbFo3fffh9f8h7i/otBlut1ykE7uT3TWK9wH1O5RrneIl/FO9eva+pv5
  Aei19f5qdpqI4IzCx0uDMeJTcekCgYEAwHF3FrdODQOh/re37eoFrk4CbQbp85J/
  cop4wzQHyiLTxqB3e1XETDa7cHPttKTX1AVMEutF5k2+y7Vd/I/MUi1HW9t/psK1
  mXjFzNAdZmpzIYHrYA2anx9lrReD0oueEExzzfcYOfruUa8udXoKqo4QvkPZVkJ/
  67wVLyw9+gMCgYEAvoxB2na+2GoVbPhyZe22i894Scjln3Sm7Mj/5Dhef61gCYwa
  pUWKbrTuVVxOPk5zF0XK5cE3rKop68zs7n5na+fMg0E7hsbzKOZ9NSJnl+za3cV1
  l5IyT0eyuxvJ61A4BJLc5sfaaF8NRZR7F3w/LanAUtq6+25XaoUl9I4PvwECgYEA
  jux2Fr/iztWQ3V1S0/aHa5HySUjmPgjicI4Y7FjbJCvDfvQ0aLwlArlvcjAXLZ9z
  z7pzamWjz0yUVDSJ7gZaJ/oK0lTttEtNlgLVXKx/+U073nnf9sGDwYQO/oPFWnxo
  0xAEvcYzDvSnRLFHXuZZv5utIbHAW0keOlTAov1HtkMCgYA34EftvNMyq4up3h3H
  El0Vx+HPWmn3Ze7wFSpgWf+ZwpCNQOpS9te06DlFSfWDk7c5VFIwFDpd9FgTPEGl
  7dONidHCAyyjLjQzj2VHT6Ph0Hs5PwlCzdHww/Q48HcWKvZy9cqNGAJRQ2p8IlrI
  1lOFqJmY0YtbcBp4Rw/nxEg8qg==
  -----END PRIVATE KEY-----`
	var runner Runner
	err := yaml.Unmarshal([]byte(input), &runner)
	require.NoError(t, err)
	assert.Equal(t, "runner-id", runner.ID)
	assert.Equal(t, "client-id", runner.ClientID)
	assert.Equal(t, "client-secret", runner.ClientSecret)
	assert.Equal(t, "webhook-secret", runner.WebhookSecret)
	assert.Equal(t, "https://example.com", runner.Host.String())
	assert.NotNil(t, runner.PrivateKey)
}
