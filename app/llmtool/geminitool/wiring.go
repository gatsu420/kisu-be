package geminitool

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/genai"
)

func (w *wiringImpl) Declare() []*genai.FunctionDeclaration {
	funcs := []*genai.FunctionDeclaration{}
	for _, t := range w.tools {
		funcs = append(funcs, t.Declaration)
	}
	return funcs
}

func (w *wiringImpl) Add(callbacks []WiringItem) {
	for _, c := range callbacks {
		w.tools[c.Declaration.Name] = c
	}
}

func (w *wiringImpl) Call(ctx context.Context, name string, token *oauth2.Token, args json.RawMessage) (json.RawMessage, error) {
	t, ok := w.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool is not registered: %v", name)
	}

	return t.Func(ctx, token, args)
}
