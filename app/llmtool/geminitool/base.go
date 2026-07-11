package geminitool

import (
	"context"
	"encoding/json"

	"github.com/gatsu420/kisu-be/app/repository/bqrepo"
	"golang.org/x/oauth2"
	"google.golang.org/genai"
)

type Tool interface {
	GetUser() WiringItem
	GetSeller() WiringItem
}

type toolImpl struct {
	bqRepo bqrepo.Repository
}

func NewTool(bqRepo bqrepo.Repository) Tool {
	return &toolImpl{
		bqRepo: bqRepo,
	}
}

type WiringItem struct {
	Declaration *genai.FunctionDeclaration
	Func        func(ctx context.Context, token *oauth2.Token, args json.RawMessage) (json.RawMessage, error)
}

type Wiring interface {
	Declare() []*genai.FunctionDeclaration
	Add(callbacks []WiringItem)
	Call(ctx context.Context, name string, token *oauth2.Token, args json.RawMessage) (json.RawMessage, error)
}

type wiringImpl struct {
	token *oauth2.Token
	tools map[string]WiringItem
}

func NewWiring() Wiring {
	return &wiringImpl{
		tools: map[string]WiringItem{},
	}
}
