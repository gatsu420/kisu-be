package bqrepo

import (
	"context"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
)

type Repository interface {
	GetUser(ctx context.Context, args GetUserArgs) ([]GetUserItem, error)
}

type repositoryImpl struct {
	projectID      string
	stringSaltPart string
	googleAuth     googleauthadapter.Adapter
}

func NewRepository(projectID string, stringSaltPart string, googleAuth googleauthadapter.Adapter) Repository {
	return &repositoryImpl{
		projectID:      projectID,
		stringSaltPart: stringSaltPart,
		googleAuth:     googleAuth,
	}
}
