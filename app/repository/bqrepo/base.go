package bqrepo

import (
	"context"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
)

type Repository interface {
	CallTool(ctx context.Context, args CallToolArgs) (CallToolResult, error)
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
