package bqrepo

import (
	"context"

	"cloud.google.com/go/bigquery"
	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
)

type Repository interface {
	GetSeller(ctx context.Context, args GetSellerArgs) ([]map[string]bigquery.Value, error)
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
