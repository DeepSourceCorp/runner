package oauth

import (
	"context"

	"github.com/deepsourcecorp/runner/auth/common"
	"golang.org/x/oauth2"
)

type Provider interface {
	AuthorizationURL(state string, scopes []string) string
	GetToken(ctx context.Context, code string) (*oauth2.Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error)
	GetUser(ctx context.Context, token *oauth2.Token) (*common.User, error)
}
