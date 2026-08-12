package pgrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	InsertUser(ctx context.Context, args InsertUserArgs) (InsertUserResult, error)
	InsertUserToken(ctx context.Context, args InsertUserTokenArgs) error
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
