package geminiadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"github.com/gatsu420/kisu-be/common/commonctx"
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
	funcDeclarations, err := a.getFuncDeclaration(ctx, getFuncDeclarationArgs{
		userID: args.UserID,
	})
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to construct function declarations: %w", err)
	}

	geminiTools := []*genai.Tool{
		{FunctionDeclarations: funcDeclarations.declarations},
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

	paramName, ok := ctx.Value(commonctx.FilterCtxKey).(string)
	if !ok {
		return GetContentResult{}, errors.New("there is no param name inside context")
	}

	contents := genai.Text(fmt.Sprintf(`
		Put %v in hashed_%v func call args.
		Translate %v into SQL.
		Strive for single tool call.
	`, args.Param, paramName, args.Prompt))
	resp, err := a.genaiClient.Models.GenerateContent(ctx, "gemini-3.1-flash-lite", contents, geminiConfig)
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to use gemini client: %w", err)
	}

	if len(resp.FunctionCalls()) == 0 {
		return GetContentResult{}, errors.New("prompt is not associated with any tool")
	}

	funcCall := resp.FunctionCalls()[0]
	funcCallParam, ok := funcCall.Args["hashed_"+paramName]
	if !ok {
		return GetContentResult{}, errors.New("there is no hashed param key inside func call args")
	}

	stringifiedFuncCallParam, ok := funcCallParam.(string)
	if !ok {
		return GetContentResult{}, errors.New("unable to cast func call param to string")
	}

	funcCallQuery, ok := funcCall.Args["query"]
	if !ok {
		return GetContentResult{}, errors.New("there is no query key inside func call args")
	}

	stringifiedFuncCallQuery, ok := funcCallQuery.(string)
	if !ok {
		return GetContentResult{}, errors.New("unable to cast func call query to string")
	}

	funcCall.Args["query"] = stringifiedFuncCallQuery +
		fmt.Sprintf(" where %v in (%v)",
			"hashed_"+paramName,
			stringifiedFuncCallParam)

	marshaledFuncCall, err := json.MarshalIndent(funcCall, "", " ")
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to marshal tool: %w", err)
	}
	stringifiedFuncCalls := string(marshaledFuncCall)
	print(stringifiedFuncCalls)

	funcCallArgs, err := json.Marshal(funcCall.Args)
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to marshal tool args: %w", err)
	}

	toolResult, err := a.metadataUsecase.CallTool(ctx, metadata.CallToolArgs{
		TableName:   funcCall.Name,
		RawToolArgs: funcCallArgs,
		Token:       args.Token,
	})
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to call tool: %w", err)
	}

	return GetContentResult{
		Content:              toolResult.Result,
		StringifiedFuncCalls: stringifiedFuncCalls,
	}, nil
}

type getFuncDeclarationArgs struct {
	userID string
}

type getFuncDeclarationResult struct {
	declarations []*genai.FunctionDeclaration
}

func (a *adapterImpl) getFuncDeclaration(ctx context.Context, args getFuncDeclarationArgs) (getFuncDeclarationResult, error) {
	tools, err := a.metadataUsecase.GetTool(ctx, metadata.GetToolArgs{
		UserID: args.userID,
	})
	if err != nil {
		return getFuncDeclarationResult{}, fmt.Errorf("unable to get tools: %w", err)
	}

	funcDeclarations := []*genai.FunctionDeclaration{}
	for _, r := range tools.Rows {
		columns := []string{}
		for _, c := range r.Columns {
			columns = append(columns, fmt.Sprintf("- %v (%v): %v",
				c.Name, c.Type, c.Description))
		}

		queryExamples := []string{}
		for _, qe := range r.QueryExamples {
			query := strings.ReplaceAll(qe.Query,
				r.TableName,
				fmt.Sprintf("%v_hashed_filter", r.TableName))
			queryExamples = append(queryExamples, fmt.Sprintf("- %v\n\t%v",
				qe.Description, query))
		}

		funcDeclarations = append(funcDeclarations, &genai.FunctionDeclaration{
			Name: r.TableName,
			Description: fmt.Sprintf(`
				Run select-only query from %v_hashed_filter to get information about: %v.

				The view has these columns:
				%v

				Sample query using the view:
				%v
				`,
				r.TableName,
				r.ToolDescription,
				strings.Join(columns, "\n"),
				strings.Join(queryExamples, "\n")),
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"hashed_" + r.ParamName: {
						Type:        genai.TypeString,
						Description: "Hashed param delimited by comma. Each element is surrounded by quote.",
					},
					"query": {
						Type:        genai.TypeString,
						Description: "Query to get wanted information without WHERE",
					},
				},
				Required: []string{"hashed_" + r.ParamName, "query"},
			},
			Response: &genai.Schema{
				Type:        genai.TypeArray,
				Items:       &genai.Schema{Type: genai.TypeString},
				Description: "List of information returned by query, 1 item represents 1 row",
			},
		})
	}

	return getFuncDeclarationResult{
		declarations: funcDeclarations,
	}, nil
}
