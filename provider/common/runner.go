package common

import "github.com/deepsourcecorp/runner/internal/signer"

type Runner struct {
	ID            string
	WebhookSecret string
}

func (r *Runner) SignPayload(payload []byte) (string, error) {
	signer, err := signer.NewSHA256Signer([]byte(r.WebhookSecret))
	if err != nil {
		return "", err
	}
	return signer.Sign(payload)
}
