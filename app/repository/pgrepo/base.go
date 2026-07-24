package pgrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	AddMetadata(ctx context.Context, args AddMetadataArgs) error
}

type repositoryImpl struct {
	pgxPool pgxpool.Pool
}

func NewRepository(pgxPool pgxpool.Pool) Repository {
	return &repositoryImpl{
		pgxPool: pgxPool,
	}
}
