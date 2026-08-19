package pgrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	AddUser(ctx context.Context, args AddUserArgs) (AddUserResult, error)
	AddUserToken(ctx context.Context, args AddUserTokenArgs) error
	GetUserToken(ctx context.Context, args GetUserTokenArgs) (GetUserTokenResult, error)
	AddTool(ctx context.Context, args AddToolArgs) error
}

type repositoryImpl struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repositoryImpl{
		pool: pool,
	}
}
