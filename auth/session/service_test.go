package session

import (
	"net/url"
	"testing"

	"github.com/deepsourcecorp/runner/auth/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_DeepSourceCallBackURL(t *testing.T) {
	baseURL, _ := url.Parse("https://deepsource.io")
	s := &Service{
		DeepSource: &common.DeepSource{
			BaseURL: baseURL,
		},
	}

	got := s.DeepSourceCallBackURL("app-id-1", url.Values{
		"key": []string{"value"},
	})

	u, err := url.Parse(got)
	require.NoError(t, err)
	assert.Equal(t, "/accounts/runner/apps/app-id-1/login/callback/bifrost/", u.Path)
	assert.Equal(t, "value", u.Query().Get("key"))
	assert.Equal(t, "https://deepsource.io", u.Scheme+"://"+u.Host)
}
