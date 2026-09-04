package geminiadapter

import (
	"context"

	"github.com/gatsu420/kisu-be/app/llmtool/geminitool"
	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"google.golang.org/genai"
)

type Adapter interface {
	GetContent(ctx context.Context, args GetContentArgs) (GetContentResult, error)
}

type adapterImpl struct {
	genaiClient      *genai.Client
	metadataUsecase  metadata.Usecase
	geminiToolWiring geminitool.Wiring
}

func NewAdapter(genaiClient *genai.Client, metadataUsecase metadata.Usecase, geminiToolWiring geminitool.Wiring) Adapter {
	return &adapterImpl{
		genaiClient:      genaiClient,
		metadataUsecase:  metadataUsecase,
		geminiToolWiring: geminiToolWiring,
	}
}
