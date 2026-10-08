package pgrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	AddAuthState(ctx context.Context, args AddAuthStateArgs) error
	ConsumeAuthState(ctx context.Context, args ConsumeAuthStateArgs) (ConsumeAuthStateResult, error)
	AddUser(ctx context.Context, args AddUserArgs) (AddUserResult, error)
	AddUserToken(ctx context.Context, args AddUserTokenArgs) error
	GetUserToken(ctx context.Context, args GetUserTokenArgs) (GetUserTokenResult, error)
	AddTool(ctx context.Context, args AddToolArgs) error
	GetTool(ctx context.Context, args GetToolArgs) (GetToolResult, error)
	AddBookmark(ctx context.Context, args AddBookmarkArgs) error
	ListBookmark(ctx context.Context, args ListBookmarkArgs) (ListBookmarkResult, error)
	GetBookmark(ctx context.Context, args GetBookmarkArgs) (GetBookmarkResult, error)
	DeleteBookmark(ctx context.Context, args DeleteBookmarkArgs) error
}

type repositoryImpl struct {
	projectID string
	pool      *pgxpool.Pool
}

func NewRepository(projectID string, pool *pgxpool.Pool) Repository {
	return &repositoryImpl{
		projectID: projectID,
		pool:      pool,
	}
}
