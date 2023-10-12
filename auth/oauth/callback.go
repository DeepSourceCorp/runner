package oauth

import (
	"fmt"
	"net/url"
)

func CallBackURL(appID string, base url.URL) *url.URL {
	path := fmt.Sprintf("/apps/%s/auth/callback", appID)
	return base.JoinPath(path)
}
