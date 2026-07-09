package googleauthadapter

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/bigquery/v2"
)

type Adapter interface {
	GetPermissionLink(state string) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	Client(ctx context.Context, token *oauth2.Token) *http.Client
}

type adapterImpl struct {
	oauthConfig *oauth2.Config
}

func NewAdapter(googleAuthClientID string, googleAuthClientSecret string, googleAuthRedirectUrl string) Adapter {
	return &adapterImpl{
		oauthConfig: &oauth2.Config{
			ClientID:     googleAuthClientID,
			ClientSecret: googleAuthClientSecret,
			RedirectURL:  googleAuthRedirectUrl,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				bigquery.BigqueryScope,
			},
			Endpoint: google.Endpoint,
		},
	}
}
