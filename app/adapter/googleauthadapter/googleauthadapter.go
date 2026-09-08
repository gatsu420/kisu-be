package googleauthadapter

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

type GetPermissionLinkArgs struct {
	State string
}

type GetPermissionLinkResult struct {
	Link string
}

func (a *adapterImpl) GetPermissionLink(args GetPermissionLinkArgs) GetPermissionLinkResult {
	link := a.oauthConfig.AuthCodeURL(args.State,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"))

	return GetPermissionLinkResult{
		Link: link,
	}
}

type ExchangeArgs struct {
	Code string
}

type ExchangeResult struct {
	Token *oauth2.Token
}

func (a *adapterImpl) Exchange(ctx context.Context, args ExchangeArgs) (ExchangeResult, error) {
	token, err := a.oauthConfig.Exchange(ctx, args.Code)

	return ExchangeResult{
		Token: token,
	}, err
}

type ClientArgs struct {
	Token *oauth2.Token
}

type ClientResult struct {
	Client *http.Client
}

func (a *adapterImpl) Client(ctx context.Context, args ClientArgs) ClientResult {
	client := a.oauthConfig.Client(ctx, args.Token)

	return ClientResult{
		Client: client,
	}
}

type TokenSourceArgs struct {
	Token *oauth2.Token
}

type TokenSourceResult struct {
	Source oauth2.TokenSource
}

func (a *adapterImpl) TokenSource(ctx context.Context, args TokenSourceArgs) TokenSourceResult {
	source := a.oauthConfig.TokenSource(ctx, args.Token)

	return TokenSourceResult{
		Source: source,
	}
}
