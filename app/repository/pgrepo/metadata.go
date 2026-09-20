package pgrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/gatsu420/kisu-be/common/commontype"
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
	Type             commontype.ToolType
	Examples         []AddToolExample
	ParamName        string
	ParamType        string
	ParamDescription string
}

type AddToolColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type AddToolExample struct {
	Description string
	Query       string
}

func (r *repositoryImpl) AddTool(ctx context.Context, args AddToolArgs) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to begin transaction for adding tool: %w", err)
	}
	defer tx.Rollback(ctx)

	var toolID string
	err = tx.QueryRow(ctx, `
		insert into tool (
			user_id, tool_description, table_name, columns, type,
			param_name, param_type, param_description
		) values ($1, $2, $3, $4, $5, $6, $7, $8)
		returning id
	`, args.UserID, args.ToolDescription, args.TableName, args.Columns, args.Type,
		args.ParamName, args.ParamType, args.ParamDescription).
		Scan(&toolID)
	if err != nil {
		return fmt.Errorf("unable to insert to tool while in transaction: %w", err)
	}

	exampleQuery := make([]string, len(args.Examples))
	exampleDescription := make([]string, len(args.Examples))
	for i, e := range args.Examples {
		exampleQuery[i] = e.Query
		exampleDescription[i] = e.Description
	}
	_, err = tx.Exec(ctx, `
		insert into example (
			tool_id, query, description
		)
		select $1, q, d
		from unnest($2::text[], $3::text[]) as t(q, d)
	`, toolID, exampleQuery, exampleDescription)
	if err != nil {
		return fmt.Errorf("unable to insert to example while in transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("unable to commit transaction for adding tool: %w", err)
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
	Type             commontype.ToolType
	Examples         []GetToolExample
	ParamName        string
	ParamType        string
	ParamDescription string
}

type GetToolColumn struct {
	Name        string
	Type        string
	Description string
}

type GetToolExample struct {
	Description string
	Query       string
}

func (r *repositoryImpl) GetTool(ctx context.Context, args GetToolArgs) (GetToolResult, error) {
	rows, err := r.pool.Query(ctx, `
		with tool_example as (
			select
				tool_id,
				jsonb_agg(
					jsonb_build_object(
						'description', description,
						'query', query
					)
				) example
			from example
			group by 1
		)

		, breakdown as (
			select
				t.tool_description,
				t.table_name,
				t.columns,
				t.type,
				coalesce(te.example, '[]'::jsonb) example,
				t.param_name,
				t.param_type,
				t.param_description
			from tool t
			left join tool_example te on
				t.id = te.tool_id
			where t.user_id = $1
		)

		select * from breakdown
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
			&resultRow.Type,
			&resultRow.Examples,
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
