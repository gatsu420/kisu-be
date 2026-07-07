package geminiadapter

import (
	"context"

	"github.com/gatsu420/kisu-be/app/llmtool/geminitool"
	"google.golang.org/genai"
)

type Adapter interface {
	GetContent(ctx context.Context, args GetContentArgs) (GetContentResult, error)
}

type adapterImpl struct {
	genaiClient      *genai.Client
	geminiToolWiring geminitool.Wiring
}

func NewAdapter(genaiClient *genai.Client, geminiToolWiring geminitool.Wiring) Adapter {
	return &adapterImpl{
		genaiClient:      genaiClient,
		geminiToolWiring: geminiToolWiring,
	}
}
