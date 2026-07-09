package geminiadapter

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/genai"
)

type GetContentArgs struct {
	Token  *oauth2.Token
	Prompt string
	Param  string
}

type GetContentResult struct {
	Content json.RawMessage
}

func (a *adapterImpl) GetContent(ctx context.Context, args GetContentArgs) (GetContentResult, error) {
	funcDeclarations := a.geminiToolWiring.Declare()
	if len(funcDeclarations) == 0 {
		return GetContentResult{}, fmt.Errorf("no tool registered")
	}

	geminiTools := []*genai.Tool{
		{FunctionDeclarations: funcDeclarations},
	}

	geminiTemp := float32(0.5)
	geminiConfig := &genai.GenerateContentConfig{
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: true,
			ThinkingLevel:   genai.ThinkingLevelMinimal,
		},
		Tools:       geminiTools,
		Temperature: &geminiTemp,
	}

	contents := genai.Text(fmt.Sprintf("Param: %v. Prompt: %v. %v",
		args.Param,
		args.Prompt,
		"Strive for single tool call, then multiple tool calls. If no relevant tool is found, dont call any tool."))

	resp, err := a.genaiClient.Models.GenerateContent(ctx, "gemini-3.1-flash-lite", contents, geminiConfig)
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to use gemini client: %w", err)
	}

	funcCalls := resp.FunctionCalls()
	if len(funcCalls) == 0 {
		return GetContentResult{
			Content: json.RawMessage("\"prompt is not associated with any tool\""),
		}, nil
	}

	funcCallArgs, err := json.Marshal(funcCalls[0].Args)
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to marshal tool args: %w", err)
	}

	result, err := a.geminiToolWiring.Call(ctx, funcCalls[0].Name, args.Token, funcCallArgs)
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to call tool: %w", err)
	}

	return GetContentResult{
		Content: result,
	}, nil
}
