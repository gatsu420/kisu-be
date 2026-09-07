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

type AddToolArgs struct {
	UserID           string
	ToolDescription  string
	TableName        string
	Columns          []AddToolColumn
	QueryExamples    []AddToolQueryExample
	ParamName        string
	ParamType        string
	ParamDescription string
}

type AddToolColumn struct {
	Name        string
	Type        string
	Description string
}

type AddToolQueryExample struct {
	Description string
	Query       string
}

func (r *repositoryImpl) AddTool(ctx context.Context, args AddToolArgs) error {
	trx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to create transaction: %w", err)
	}
	defer trx.Rollback(ctx)

	var toolID string
	err = trx.QueryRow(ctx, `
		insert into tool (
			user_id, tool_description, table_name, columns, query_examples
		) values ($1, $2, $3, $4, $5)
		returning id
	`, args.UserID, args.ToolDescription, args.TableName, args.Columns, args.QueryExamples).
		Scan(&toolID)
	if err != nil {
		return fmt.Errorf("unable to add tool: %w", err)
	}

	_, err = trx.Exec(ctx, `
		insert into tool_param (
			tool_id, name, type, description
		) values ($1, $2, $3, $4)
	`, toolID, args.ParamName, args.ParamType, args.ParamDescription)
	if err != nil {
		return fmt.Errorf("unable to add tool_param when adding tool: %w", err)
	}

	err = trx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("unable to commit transaction when adding tool: %w", err)
	}

	return nil
}

type GetToolArgs struct {
	UserID string
}

type GetToolResult struct {
	Rows []GetToolRow
}

type GetToolRow struct {
	ToolDescription  string
	TableName        string
	Columns          []GetToolColumn
	QueryExamples    []GetToolQueryExample
	ParamName        string
	ParamType        string
	ParamDescription string
}

type GetToolColumn struct {
	Name        string
	Type        string
	Description string
}

type GetToolQueryExample struct {
	Description string
	Query       string
}

func (r *repositoryImpl) GetTool(ctx context.Context, args GetToolArgs) (GetToolResult, error) {
	rows, err := r.pool.Query(ctx, `
		select
			t.tool_description,
			t.table_name,
			t.columns,
			t.query_examples,
			tp.name param_name,
			tp.type param_type,
			tp.description param_description
		from tool t
		left join tool_param tp on
			t.id = tp.tool_id
		where t.user_id = $1
	`, args.UserID)
	if err != nil {
		return GetToolResult{}, fmt.Errorf("unable to get tool: %w", err)
	}
	defer rows.Close()

	resultRows := []GetToolRow{}
	for rows.Next() {
		var resultRow GetToolRow
		err := rows.Scan(
			&resultRow.ToolDescription,
			&resultRow.TableName,
			&resultRow.Columns,
			&resultRow.QueryExamples,
			&resultRow.ParamName,
			&resultRow.ParamType,
			&resultRow.ParamDescription,
		)
		if err != nil {
			return GetToolResult{}, fmt.Errorf("unable to read row when getting tool: %w", err)
		}

		resultRows = append(resultRows, resultRow)
	}

	err = rows.Err()
	if err != nil {
		return GetToolResult{}, fmt.Errorf("unable to iterate rows when getting tool: %w", err)
	}

	return GetToolResult{
		Rows: resultRows,
	}, nil
}
