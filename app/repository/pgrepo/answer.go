package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gatsu420/kisu/common/commonerr"
	"github.com/jackc/pgx/v5"
)

type AddBookmarkArgs struct {
	ID         string
	UserID     string
	Name       string
	ParamName  string
	ParamValue string
	Query      string
	HashedTool string
}

func (r *repositoryImpl) AddBookmark(ctx context.Context, args AddBookmarkArgs) error {
	nameRunes := []rune(args.Name)
	if len(nameRunes) > 30 {
		args.Name = string(nameRunes[:30]) + "..."
	}

	_, err := r.pool.Exec(ctx, `
		insert into bookmark (
			id, user_id, name,
			param_name, param_value, query, hashed_tool
		) values ($1, $2, $3, $4, $5, $6, $7)
		on conflict (id) do update set
			name = excluded.name,
			param_name = excluded.param_name,
			param_value = excluded.param_value,
			query = excluded.query,
			hashed_tool = excluded.hashed_tool,
			updated_at = now()
		where bookmark.user_id = excluded.user_id
		`, args.ID, args.UserID, args.Name,
		args.ParamName, args.ParamValue, args.Query, args.HashedTool)
	if err != nil {
		return fmt.Errorf("unable to add bookmark: %w", err)
	}

	return nil
}

type ListBookmarkArgs struct {
	UserID string
}

type ListBookmarkResult struct {
	Rows []ListBookmarkRow
}

type ListBookmarkRow struct {
	ID        string
	Name      string
	UpdatedAt time.Time
}

func (r *repositoryImpl) ListBookmark(ctx context.Context, args ListBookmarkArgs) (ListBookmarkResult, error) {
	rows, err := r.pool.Query(ctx, `
		select
			id,
			name,
			updated_at
		from bookmark
		where user_id = $1
	`, args.UserID)
	if err != nil {
		return ListBookmarkResult{}, fmt.Errorf("unable to list bookmark: %w", err)
	}
	defer rows.Close()

	resultRows := []ListBookmarkRow{}
	for rows.Next() {
		var resultRow ListBookmarkRow
		err := rows.Scan(&resultRow.ID, &resultRow.Name, &resultRow.UpdatedAt)
		if err != nil {
			return ListBookmarkResult{}, fmt.Errorf("unable to read row when listing bookmark: %w", err)
		}

		resultRows = append(resultRows, resultRow)
	}

	if rows.Err() != nil {
		return ListBookmarkResult{}, fmt.Errorf("unable to iterate rows when listing bookmark: %w", rows.Err())
	}

	return ListBookmarkResult{
		Rows: resultRows,
	}, nil
}

type GetBookmarkArgs struct {
	ID     string
	UserID string
}

type GetBookmarkResult struct {
	Name       string
	ParamName  string
	ParamValue string
	Query      string
	HashedTool string
	UpdatedAt  time.Time
}

func (r *repositoryImpl) GetBookmark(ctx context.Context, args GetBookmarkArgs) (GetBookmarkResult, error) {
	var result GetBookmarkResult
	err := r.pool.QueryRow(ctx, `
		select
			name,
			param_name,
			param_value,
			query,
			hashed_tool,
			updated_at
		from bookmark
		where id = $1
		and user_id = $2
	`, args.ID, args.UserID).
		Scan(&result.Name,
			&result.ParamName,
			&result.ParamValue,
			&result.Query,
			&result.HashedTool,
			&result.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetBookmarkResult{}, commonerr.NoRowsQueryErr
		}

		return GetBookmarkResult{}, fmt.Errorf("unable to get bookmark: %w", err)
	}

	return result, nil
}

type DeleteBookmarkArgs struct {
	ID string
}

func (r *repositoryImpl) DeleteBookmark(ctx context.Context, args DeleteBookmarkArgs) error {
	_, err := r.pool.Exec(ctx, `
		delete from bookmark
		where id = $1
	`, args.ID)
	if err != nil {
		return fmt.Errorf("unable to delete bookmark: %w", err)
	}

	return nil
}
