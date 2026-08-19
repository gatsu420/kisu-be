package pgrepo

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

type AddUserArgs struct {
	Email string
}

type AddUserResult struct {
	UserID string
}

func (r *repositoryImpl) AddUser(ctx context.Context, args AddUserArgs) (AddUserResult, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `
		insert into user_information (email)
		values ($1)
		on conflict (email) do update set
			email = excluded.email
		returning id
	`, args.Email).Scan(&userID)
	if err != nil {
		return AddUserResult{}, fmt.Errorf("unable to add user: %w", err)
	}

	return AddUserResult{
		UserID: userID,
	}, nil
}

type AddUserTokenArgs struct {
	UserID string
	Token  *oauth2.Token
}

func (r *repositoryImpl) AddUserToken(ctx context.Context, args AddUserTokenArgs) error {
	_, err := r.pool.Exec(ctx, `
		insert into user_token (user_id, access_token, refresh_token, expired_at)
		values ($1, $2, $3, $4)
		on conflict (user_id) do update set
			access_token = $2,
			refresh_token = $3,
			expired_at = $4,
			updated_at = now()
	`, args.UserID, args.Token.AccessToken, args.Token.RefreshToken, args.Token.Expiry)
	if err != nil {
		return fmt.Errorf("unable to add user token: %w", err)
	}

	return nil
}

type GetUserTokenArgs struct {
	UserID string
}

type GetUserTokenResult struct {
	Token *oauth2.Token
}

func (r *repositoryImpl) GetUserToken(ctx context.Context, args GetUserTokenArgs) (GetUserTokenResult, error) {
	var accessToken string
	var refreshToken string
	var expiredAt time.Time

	err := r.pool.QueryRow(ctx, `
		select
			access_token,
			refresh_token,
			expired_at
		from user_token
		where user_id = $1
	`, args.UserID).Scan(&accessToken, &refreshToken, &expiredAt)
	if err != nil {
		return GetUserTokenResult{}, fmt.Errorf("unable to get user token: %w", err)
	}

	return GetUserTokenResult{
		Token: &oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Expiry:       expiredAt,
		},
	}, nil
}

type AddToolColumn struct {
	Name        string
	Type        string
	Description string
	IsSelected  bool
}

type AddToolQueryExample struct {
	Description string
	Query       string
}

type AddToolArgs struct {
	ToolDescription string
	TableName       string
	Columns         []AddToolColumn
	QueryExamples   []AddToolQueryExample
}

func (r *repositoryImpl) AddTool(ctx context.Context, args AddToolArgs) error {
	_, err := r.pool.Exec(ctx, `
		insert into tool (
			tool_description, table_name, columns, query_examples
		) values ($1, $2, $3, $4)
	`, args.ToolDescription, args.TableName, args.Columns, args.QueryExamples)
	if err != nil {
		return fmt.Errorf("unable to add tool: %w", err)
	}

	return nil
}
