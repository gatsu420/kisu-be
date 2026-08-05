package usertokenrepo

import (
	"context"
	"database/sql"

	"golang.org/x/oauth2"
)

type Repository interface {
	Get(ctx context.Context, userID string) (*oauth2.Token, error)
	Save(ctx context.Context, userID string, token *oauth2.Token) error
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repositoryImpl{
		db: db,
	}
}
