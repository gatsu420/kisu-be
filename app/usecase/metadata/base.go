package metadata

import (
	"context"

	"github.com/gatsu420/kisu-be/app/repository/bqrepo"
	"github.com/gatsu420/kisu-be/app/repository/pgrepo"
)

type Usecase interface {
	AddUser(ctx context.Context, args AddUserArgs) (AddUserResult, error)
	AddUserToken(ctx context.Context, args AddUserTokenArgs) error
	GetUserToken(ctx context.Context, args GetUserTokenArgs) (GetUserTokenResult, error)
	AddTool(ctx context.Context, args AddToolArgs) error
	GetTool(ctx context.Context, args GetToolArgs) (GetToolResult, error)
	CallTool(ctx context.Context, args CallToolArgs) (CallToolResult, error)
}

type usecaseImpl struct {
	pgRepo pgrepo.Repository
	bqRepo bqrepo.Repository
}

func NewUsecase(pgRepo pgrepo.Repository, bqRepo bqrepo.Repository) Usecase {
	return &usecaseImpl{
		pgRepo: pgRepo,
		bqRepo: bqRepo,
	}
}
