package pgrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/gatsu420/kisu-be/common/commontype"
	"golang.org/x/oauth2"
)

type AddAuthStateArgs struct {
	State string
}

func (r *repositoryImpl) AddAuthState(ctx context.Context, args AddAuthStateArgs) error {
	_, err := r.pool.Exec(ctx, `
		insert into auth_state (state)
		values ($1)
	`, args.State)
	if err != nil {
		return fmt.Errorf("unable to add auth state: %w", err)
	}

	return nil
}

type ConsumeAuthStateArgs struct {
	State string
}

type ConsumeAuthStateResult struct {
	StateExistence bool
}

func (r *repositoryImpl) ConsumeAuthState(ctx context.Context, args ConsumeAuthStateArgs) (ConsumeAuthStateResult, error) {
	var state string
	err := r.pool.QueryRow(ctx, `
		delete from auth_state
		where state = $1
		returning state
	`, args.State).Scan(&state)
	if err != nil {
		return ConsumeAuthStateResult{}, fmt.Errorf("unable to delete from auth_state table: %w", err)
	}

	return ConsumeAuthStateResult{
		StateExistence: state != "",
	}, nil
}

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
	UserID          string
	ToolDescription string
	Project         string
	Dataset         string
	TableName       string
	Columns         []AddToolColumn
	Type            commontype.ToolType
	Examples        []AddToolExample
	ParamNames      []string
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

	var query string
	queryArgs := []any{}
	switch args.Type {
	case commontype.TableToolType:
		query = `
			insert into tool (
				user_id, tool_description, project, dataset, table_name,
				columns, type
			) values ($1, $2, $3, $4, $5, $6, $7)
			returning id
		`
		queryArgs = append(queryArgs,
			args.UserID, args.ToolDescription, args.Project, args.Dataset, args.TableName,
			args.Columns, args.Type)

	case commontype.QueryToolType:
		query = `
			insert into tool (
				user_id, tool_description, project,
				columns, type
			) values ($1, $2, $3, $4, $5)
			returning id
		`
		queryArgs = append(queryArgs,
			args.UserID, args.ToolDescription, r.projectID,
			args.Columns, args.Type)
	}

	var toolID string
	err = tx.QueryRow(ctx, query, queryArgs...).
		Scan(&toolID)
	if err != nil {
		return fmt.Errorf("unable to insert to tool while in transaction: %w", err)
	}

	_, err = tx.Exec(ctx, `
		insert into tool_param (tool_id, name)
		select $1, n
		from unnest($2::text[]) as t(n)
	`, toolID, args.ParamNames)
	if err != nil {
		return fmt.Errorf("unable to insert to tool_param while in transaction: %w", err)
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
	ID              string
	ToolDescription string
	Project         string
	Dataset         *string
	TableName       *string
	Columns         []GetToolColumn
	Type            commontype.ToolType
	Examples        []GetToolExample
	ParamNames      []string
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

		, tool_param as (
			select
				tool_id,
				jsonb_agg(name) name
			from tool_param
			group by 1
		)

		, breakdown as (
			select
				t.id,
				t.tool_description,
				t.project,
				t.dataset,
				t.table_name,
				t.columns,
				t.type,
				coalesce(te.example, '[]'::jsonb) example,
				coalesce(tp.name, '[]'::jsonb) param_name
			from tool t
			left join tool_example te on
				t.id = te.tool_id
			left join tool_param tp on
				t.id = tp.tool_id
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
			&resultRow.ID,
			&resultRow.ToolDescription,
			&resultRow.Project,
			&resultRow.Dataset,
			&resultRow.TableName,
			&resultRow.Columns,
			&resultRow.Type,
			&resultRow.Examples,
			&resultRow.ParamNames,
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
