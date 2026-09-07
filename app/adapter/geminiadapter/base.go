package geminiadapter

import (
	"context"

	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"google.golang.org/genai"
)

type Adapter interface {
	GetContent(ctx context.Context, args GetContentArgs) (GetContentResult, error)
}

type adapterImpl struct {
	genaiClient     *genai.Client
	metadataUsecase metadata.Usecase
}

func NewAdapter(genaiClient *genai.Client, metadataUsecase metadata.Usecase) Adapter {
	return &adapterImpl{
		genaiClient:     genaiClient,
		metadataUsecase: metadataUsecase,
	}
}
