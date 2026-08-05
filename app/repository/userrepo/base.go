package userrepo

import (
	"context"
	"database/sql"
)

type Repository interface {
	UpsertByEmail(ctx context.Context, email string) (string, error)
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repositoryImpl{
		db: db,
	}
}
