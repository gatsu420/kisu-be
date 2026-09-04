package geminiadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"golang.org/x/oauth2"
	"google.golang.org/genai"
)

type GetContentArgs struct {
	Token  *oauth2.Token
	Prompt string
	Param  string
	UserID string
}

type GetContentResult struct {
	Content              json.RawMessage
	StringifiedFuncCalls string
}

func (a *adapterImpl) GetContent(ctx context.Context, args GetContentArgs) (GetContentResult, error) {
	funcDeclarations, err := a.getFuncDeclaration(ctx, args.UserID)
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to construct function declarations: %w", err)
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
	marshaledFuncCalls, err := json.MarshalIndent(funcCalls, "", " ")
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to marshal tool: %w", err)
	}
	stringifiedFuncCalls := string(marshaledFuncCalls)

	if len(funcCalls) == 0 {
		return GetContentResult{
			Content:              json.RawMessage("\"prompt is not associated with any tool\""),
			StringifiedFuncCalls: stringifiedFuncCalls,
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
		Content:              result,
		StringifiedFuncCalls: stringifiedFuncCalls,
	}, nil
}

func (a *adapterImpl) getFuncDeclaration(ctx context.Context, userID string) ([]*genai.FunctionDeclaration, error) {
	tools, err := a.metadataUsecase.GetTool(ctx, metadata.GetToolArgs{
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get tools: %w", err)
	}

	funcDeclarations := []*genai.FunctionDeclaration{}
	for _, t := range tools {
		columns := []string{}
		for _, c := range t.Columns {
			columns = append(columns, fmt.Sprintf("- %v (%v): %v",
				c.Name, c.Type, c.Description))
		}

		queryExamples := []string{}
		for _, qe := range t.QueryExample {
			queryExamples = append(queryExamples, fmt.Sprintf("- %v\n\t%v",
				qe.Description,
				strings.ReplaceAll(qe.Query,
					fmt.Sprintf("from %v", t.TableName),
					fmt.Sprintf("from %v_view", t.TableName))))
		}

		funcDeclarations = append(funcDeclarations, &genai.FunctionDeclaration{
			Name: t.TableName,
			Description: fmt.Sprintf(`
				Run select-only query from %v_view to get information about: %v.

				The view has these columns:
				%v

				Column %v doesn't need to be selected.

				Sample query using the view:
				%v
				`,
				t.TableName,
				t.ToolDescription,
				strings.Join(columns, "\n"),
				t.ParamName,
				strings.Join(queryExamples, "\n")),
		})
	}

	return funcDeclarations, nil
}
