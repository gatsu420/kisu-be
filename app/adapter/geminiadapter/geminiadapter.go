package geminiadapter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"github.com/gatsu420/kisu-be/common/commonctx"
	"github.com/gatsu420/kisu-be/common/commontype"
	"google.golang.org/genai"
)

type GetContentArgs struct {
	Prompt     string
	ParamValue string
	UserID     string
}

type GetContentResult struct {
	Tool *genai.FunctionCall
	Type commontype.ToolType
}

func (a *adapterImpl) GetContent(ctx context.Context, args GetContentArgs) (GetContentResult, error) {
	toolDeclarations, err := a.declareTool(ctx, declareToolArgs{
		userID: args.UserID,
	})
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to construct function declarations: %w", err)
	}

	geminiTools := []*genai.Tool{
		{FunctionDeclarations: toolDeclarations.declarations},
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

	paramName, ok := ctx.Value(commonctx.ParamNameCtxKey).(string)
	if !ok {
		return GetContentResult{}, errors.New("there is no param name inside context")
	}

	contents := genai.Text(fmt.Sprintf(`
		Put %s in hashed_%s tool args.
		Translate %s into SQL.
		Strive for single tool call.
	`, args.ParamValue, paramName, args.Prompt))
	resp, err := a.genaiClient.Models.GenerateContent(ctx, "gemini-3.1-flash-lite", contents, geminiConfig)
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to use gemini client: %w", err)
	}

	if len(resp.FunctionCalls()) == 0 {
		return GetContentResult{}, errors.New("prompt is not associated with any tool")
	}

	tool := resp.FunctionCalls()[0]
	return GetContentResult{
		Tool: tool,
		Type: toolDeclarations.toolTypes[tool.Name],
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

		var toolName string
		var tableLocation string
		switch r.Type {
		case commontype.TableToolType:
			tableLocation = fmt.Sprintf("%s.%s.%s",
				r.Project,
				r.Dataset,
				r.TableName)
			toolName = tableLocation
		case commontype.QueryToolType:
			toolName = "t" + r.ID
		}

		columnItems := []string{}
		columns := `
		The CTE has these columns:
		`
		for _, c := range r.Columns {
			columnItems = append(columnItems,
				fmt.Sprintf(`
				-	%s (%s)
					%s
				`,
					c.Name, c.Type, c.Description))
		}
		columns += strings.Join(columnItems, "\n")

		var builderQuery string
		if r.Type == commontype.QueryToolType {
			builderQuery = fmt.Sprintf(`
			The CTE is built using this builder query:
			%s
			`, r.Examples[0].Query)
		}

		exampleItems := []string{}
		examples := `
		Example query using the CTE:
		`
		if r.Type == commontype.TableToolType {
			for _, e := range r.Examples {
				exampleItems = append(exampleItems,
					fmt.Sprintf(`
					-	%s
						%s
					`,
						e.Description,
						strings.ReplaceAll(e.Query,
							"`"+tableLocation+"`",
							"hashed_param")))
			}

			examples += strings.Join(exampleItems, "\n")
		}

		hashedParamNames := make([]string, len(r.ParamNames))
		hashedParamProp := make(map[string]*genai.Schema, len(r.ParamNames))
		for i, pn := range r.ParamNames {
			hashedParamNames[i] = "hashed_" + pn
			hashedParamProp[pn] = &genai.Schema{
				Type:        genai.TypeString,
				Description: "Hashed param delimited by comma. Each element is surrounded by quote.",
			}
		}

		// A tool may have multiple param, but we ask gemini to choose
		// only one.
		hashedParamMaxProp := int64(1)

		declarations = append(declarations, &genai.FunctionDeclaration{
			Name: toolName,
			Description: fmt.Sprintf(`
				Run SELECT-only query from hashed_param CTE to get
				information about: %s.

				%s

				%s

				%s

				The query should be without WHERE.
				Do not SELECT %s.
				`,
				r.ToolDescription,
				columns,
				builderQuery,
				examples,
				strings.Join(hashedParamNames, ",")),

			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"hashed_param": {
						Type:          genai.TypeObject,
						Description:   "Key-value of hashed param name and its value",
						Properties:    hashedParamProp,
						MaxProperties: &hashedParamMaxProp,
					},
					"query": {
						Type:        genai.TypeString,
						Description: "Query to get wanted information",
					},
					"builder_query": {
						Type:        genai.TypeString,
						Description: "Query that is used to build hashed_param CTE",
					},
				},
				Required: []string{"hashed_param", "query", "builder_query"},
			},

			Response: &genai.Schema{
				Type:        genai.TypeArray,
				Items:       &genai.Schema{Type: genai.TypeString},
				Description: "List of information returned by query, 1 item represents 1 row",
			},
		})
		toolTypes[toolName] = r.Type
	}

	return declareToolResult{
		declarations: declarations,
		toolTypes:    toolTypes,
	}, nil
}
