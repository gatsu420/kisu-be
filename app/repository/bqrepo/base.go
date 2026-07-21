package bqrepo

import (
	"context"

	"cloud.google.com/go/bigquery"
	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
)

type Repository interface {
	GetInformation(ctx context.Context, args GetSellerArgs) ([]map[string]bigquery.Value, error)
}

type repositoryImpl struct {
	projectID  string
	googleAuth googleauthadapter.Adapter
}

func NewRepository(projectID string, googleAuth googleauthadapter.Adapter) Repository {
	return &repositoryImpl{
		projectID:  projectID,
		googleAuth: googleAuth,
	}
}
