package promptanswer

import (
	"context"

	"github.com/gatsu420/kisu-be/app/adapter/geminiadapter"
	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/app/repository/tokenrepo"
)

type Usecase interface {
	GetAnswer(ctx context.Context, args GetAnswerArgs) (GetAnswerResult, error)
}

type usecaseImpl struct {
	tokenRepo     tokenrepo.Repository
	googleAuth    googleauthadapter.Adapter
	geminiAdapter geminiadapter.Adapter
}

func NewUsecase(tokenRepo tokenrepo.Repository, googleAuth googleauthadapter.Adapter, geminiAdapter geminiadapter.Adapter) Usecase {
	return &usecaseImpl{
		tokenRepo:     tokenRepo,
		googleAuth:    googleAuth,
		geminiAdapter: geminiAdapter,
	}
}
