package googleauthadapter

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

func (a *adapterImpl) GetPermissionLink(state string) string {
	return a.oauthConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"))
}

func (a *adapterImpl) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return a.oauthConfig.Exchange(ctx, code)
}

func (a *adapterImpl) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return a.oauthConfig.Client(ctx, token)
}
