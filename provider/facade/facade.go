package facade

import (
	"net/url"

	"github.com/deepsourcecorp/runner/provider/github"
	"golang.org/x/exp/slog"
)

type ProviderFacade struct {
	GithubAPIProxyFactory *github.APIProxyFactory
}

func NewProviderFacade(githubAPIProxyFactory *github.APIProxyFactory) *ProviderFacade {
	return &ProviderFacade{
		GithubAPIProxyFactory: githubAPIProxyFactory,
	}
}

func (f *ProviderFacade) AuthenticatedRemoteURL(appID, installationID string, srcURL string) (string, error) {
	proxy, err := f.GithubAPIProxyFactory.NewProxy(appID, installationID)
	if err != nil {
		return "", err
	}
	jwt, err := proxy.GenerateJWT()
	if err != nil {
		return "", err
	}
	token, err := proxy.GenerateAccessToken(jwt)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(srcURL)
	if err != nil {
		slog.Error("failed to parse remote url, url = %s, err= %v", srcURL, err)
		return "", err
	}

	u.User = url.UserPassword("x-access-token", token)
	return u.String(), nil
}
