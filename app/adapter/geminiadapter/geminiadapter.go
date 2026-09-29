package geminiadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"github.com/gatsu420/kisu-be/common/commonctx"
	"github.com/gatsu420/kisu-be/common/commontype"
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
	Content         json.RawMessage
	StringifiedTool string
}

func (a *adapterImpl) GetContent(ctx context.Context, args GetContentArgs) (GetContentResult, error) {
	funcDeclarations, err := a.declareTool(ctx, declareToolArgs{
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
		Put %v in hashed_%v tool args.
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

	tool := resp.FunctionCalls()[0]
	toolBuilderQuery, ok := tool.Args["builder_query"]
	if !ok {
		return GetContentResult{}, errors.New("there is no builder_query key inside tool args")
	}

	stringifiedToolBuilderQuery, ok := toolBuilderQuery.(string)
	if !ok {
		return GetContentResult{}, errors.New("unable to cast tool builder query to string")
	}

	toolQuery, ok := tool.Args["query"]
	if !ok {
		return GetContentResult{}, errors.New("there is no query key inside tool args")
	}

	stringifiedToolQuery, ok := toolQuery.(string)
	if !ok {
		return GetContentResult{}, errors.New("unable to cast tool query to string")
	}

	toolParam, ok := tool.Args["hashed_"+paramName]
	if !ok {
		return GetContentResult{}, errors.New("there is no hashed param key inside tool args")
	}

	stringifiedToolParam, ok := toolParam.(string)
	if !ok {
		return GetContentResult{}, errors.New("unable to cast tool param to string")
	}

	stringifiedToolQuery += fmt.Sprintf(" where %s in (%s)",
		"hashed_"+paramName,
		stringifiedToolParam)

	toolResult, err := a.metadataUsecase.CallTool(ctx, metadata.CallToolArgs{
		Type:          funcDeclarations.toolTypes[tool.Name],
		TableLocation: tool.Name,
		BuilderQuery:  stringifiedToolBuilderQuery,
		Query:         stringifiedToolQuery,
		Token:         args.Token,
	})
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to call tool: %w", err)
	}

	marshaledTool, err := json.MarshalIndent(tool, "", " ")
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to marshal tool: %w", err)
	}

	return GetContentResult{
		Content:         toolResult.Result,
		StringifiedTool: string(marshaledTool),
	}, nil
}

type declareToolArgs struct {
	userID string
}

type declareToolResult struct {
	declarations []*genai.FunctionDeclaration
	toolTypes    map[string]commontype.ToolType
}

func (a *adapterImpl) declareTool(ctx context.Context, args declareToolArgs) (declareToolResult, error) {
	tools, err := a.metadataUsecase.GetTool(ctx, metadata.GetToolArgs{
		UserID: args.userID,
	})
	if err != nil {
		return declareToolResult{}, fmt.Errorf("unable to get tools: %w", err)
	}

	declarations := []*genai.FunctionDeclaration{}
	toolTypes := map[string]commontype.ToolType{}
	for _, r := range tools.Rows {
		if !r.Type.ValidateToolType() {
			return declareToolResult{}, errors.New("invalid tool type")
		}

		columns := []string{}
		for _, c := range r.Columns {
			columns = append(columns, fmt.Sprintf("- %v (%v): %v",
				c.Name, c.Type, c.Description))
		}

		var builderQuery string
		if r.Type == commontype.QueryToolType {
			builderQuery = fmt.Sprintf(`
			That view is build using this builder query:
			%v
			`, r.Examples[0].Query)
		}

		examples := []string{}
		for _, e := range r.Examples {
			if r.Type == commontype.TableToolType {
				queryWithHashedFilter := strings.ReplaceAll(e.Query,
					r.TableName,
					r.TableName+"_hashed_filter")
				examples = append(examples, fmt.Sprintf("- %s\n\t%s",
					e.Description, queryWithHashedFilter))
			} else {
				examples = append(examples, fmt.Sprintf("- %s\n\t%s",
					e.Description, e.Query))
			}
		}

		tableLocation := fmt.Sprintf("%s.%s.%s",
			r.Project,
			r.Dataset,
			r.TableName)
		toolTypes[tableLocation] = r.Type

		declarations = append(declarations, &genai.FunctionDeclaration{
			Name: tableLocation,
			Description: fmt.Sprintf(`
				Run select-only query from %s.%s.%s_hashed_filter to get
				information about: %s.

				%s

				The view has these columns:
				%s

				Sample query using the view:
				%s
				`, r.Project,
				r.Dataset,
				r.TableName,
				r.ToolDescription,
				builderQuery,
				strings.Join(columns, "\n"),
				strings.Join(examples, "\n")),

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
					"builder_query": {
						Type:        genai.TypeString,
						Description: "Query that is used to build *_hashed_filter view",
					},
				},
				Required: []string{"hashed_" + r.ParamName, "query", "builder_query"},
			},

			Response: &genai.Schema{
				Type:        genai.TypeArray,
				Items:       &genai.Schema{Type: genai.TypeString},
				Description: "List of information returned by query, 1 item represents 1 row",
			},
		})
	}

	return declareToolResult{
		declarations: declarations,
		toolTypes:    toolTypes,
	}, nil
}
