package auth

type Runner struct {
	ClientID     string
	ClientSecret string
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
