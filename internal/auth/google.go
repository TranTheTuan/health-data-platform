package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/TranTheTuan/health-data-platform/configs"
)

var GoogleOAuthConfig *oauth2.Config

// InitGoogleOAuth configures the oauth2.Config instance. This needs to be called before handling requests.
func InitGoogleOAuth(cfg *configs.Config) {
	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
