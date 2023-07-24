package oauth

import "fmt"

func CallbackURL(appID string) string {
	return fmt.Sprintf("/apps/%s/auth/callback", appID)
}
