package answer

import (
	"context"

	"github.com/gatsu420/kisu-be/app/adapter/geminiadapter"
)

type Usecase interface {
	GetAnswer(ctx context.Context, args GetAnswerArgs) (GetAnswerResult, error)
}

type usecaseImpl struct {
	geminiAdapter geminiadapter.Adapter
}

func NewUsecase(geminiAdapter geminiadapter.Adapter) Usecase {
	return &usecaseImpl{
		geminiAdapter: geminiAdapter,
	}
}
