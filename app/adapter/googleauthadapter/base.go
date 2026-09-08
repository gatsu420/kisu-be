package googleauthadapter

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/bigquery/v2"
)

type Adapter interface {
	GetPermissionLink(args GetPermissionLinkArgs) GetPermissionLinkResult
	Exchange(ctx context.Context, args ExchangeArgs) (ExchangeResult, error)
	Client(ctx context.Context, args ClientArgs) ClientResult
	TokenSource(ctx context.Context, args TokenSourceArgs) TokenSourceResult
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
