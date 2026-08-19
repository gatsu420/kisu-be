package metadata

import (
	"context"

	"github.com/gatsu420/kisu-be/app/repository/pgrepo"
)

type Usecase interface {
	AddUser(ctx context.Context, args AddUserArgs) (AddUserResult, error)
	AddUserToken(ctx context.Context, args AddUserTokenArgs) error
	GetUserToken(ctx context.Context, args GetUserTokenArgs) (GetUserTokenResult, error)
	AddTool(ctx context.Context, args AddToolArgs) error
}

type usecaseImpl struct {
	pgRepo pgrepo.Repository
}

func NewUsecase(pgRepo pgrepo.Repository) Usecase {
	return &usecaseImpl{
		pgRepo: pgRepo,
	}
}
