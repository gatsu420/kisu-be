package metadata

import (
	"context"

	"github.com/gatsu420/kisu-be/app/repository/pgrepo"
)

type Usecase interface {
	InsertUser(ctx context.Context, args InsertUserArgs) (InsertUserResult, error)
	InsertUserToken(ctx context.Context, args InsertUserTokenArgs) error
	GetUserToken(ctx context.Context, args GetUserTokenArgs) (GetUserTokenResult, error)
}

type usecaseImpl struct {
	pgRepo pgrepo.Repository
}

func NewUsecase(pgRepo pgrepo.Repository) Usecase {
	return &usecaseImpl{
		pgRepo: pgRepo,
	}
}
