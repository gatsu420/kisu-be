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

	paramName, ok := ctx.Value(commonctx.FilterCtxKey).(string)
	if !ok {
		return GetContentResult{}, errors.New("there is no param name inside context")
	}

	funcCall, err := a.generateFuncCall(ctx, generateFuncCallArgs{
		funcDeclarations: funcDeclarations.declarations,
		prompt:           args.Prompt,
		paramName:        paramName,
		param:            args.Param,
	})
	if err != nil {
		return GetContentResult{}, err
	}

	toolArg, err := buildToolArg(funcCall, paramName)
	if err != nil {
		return GetContentResult{}, err
	}

	marshaledFuncCall, err := json.MarshalIndent(funcCall, "", " ")
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to marshal tool: %w", err)
	}

	rawToolArgs, err := json.Marshal(funcCall.Args)
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to marshal tool args: %w", err)
	}

	toolResult, err := a.metadataUsecase.CallTool(ctx, metadata.CallToolArgs{
		Type:          funcDeclarations.toolTypes[funcCall.Name],
		TableLocation: funcCall.Name,
		Query:         toolArg.query,
		BuilderQuery:  toolArg.builderQuery,
		RawToolArgs:   rawToolArgs,
		Token:         args.Token,
	})
	if err != nil {
		return GetContentResult{}, fmt.Errorf("unable to call tool: %w", err)
	}

	return GetContentResult{
		Content:              toolResult.Result,
		StringifiedFuncCalls: string(marshaledFuncCall),
	}, nil
}

type generateFuncCallArgs struct {
	funcDeclarations []*genai.FunctionDeclaration
	prompt           string
	paramName        string
	param            string
}

func (a *adapterImpl) generateFuncCall(ctx context.Context, args generateFuncCallArgs) (*genai.FunctionCall, error) {
	config := buildGenerateContentConfig(args.funcDeclarations)
	contents := genai.Text(fmt.Sprintf(`
		Put %v in hashed_%v func call args.
		Translate %v into SQL.
		Strive for single tool call.
	`, args.param, args.paramName, args.prompt))

	resp, err := a.genaiClient.Models.GenerateContent(ctx, "gemini-3.1-flash-lite", contents, config)
	if err != nil {
		return nil, fmt.Errorf("unable to use gemini client: %w", err)
	}

	if len(resp.FunctionCalls()) == 0 {
		return nil, errors.New("prompt is not associated with any tool")
	}

	return resp.FunctionCalls()[0], nil
}

func buildGenerateContentConfig(funcDeclarations []*genai.FunctionDeclaration) *genai.GenerateContentConfig {
	geminiTemp := float32(0.5)
	return &genai.GenerateContentConfig{
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: true,
			ThinkingLevel:   genai.ThinkingLevelMinimal,
		},
		Tools: []*genai.Tool{
			{FunctionDeclarations: funcDeclarations},
		},
		Temperature: &geminiTemp,
	}
}

type toolArg struct {
	query        string
	builderQuery string
}

func buildToolArg(funcCall *genai.FunctionCall, paramName string) (toolArg, error) {
	hashedParam, err := getFuncCallStringArg(funcCall, "hashed_"+paramName, "hashed param", "param")
	if err != nil {
		return toolArg{}, err
	}

	query, err := getFuncCallStringArg(funcCall, "query", "query", "query")
	if err != nil {
		return toolArg{}, err
	}

	builderQuery, err := getFuncCallStringArg(funcCall, "builder_query", "builder_query", "builder query")
	if err != nil {
		return toolArg{}, err
	}

	funcCall.Args["query"] = query +
		fmt.Sprintf(" where %v in (%v)",
			"hashed_"+paramName,
			hashedParam)

	return toolArg{
		query:        query,
		builderQuery: builderQuery,
	}, nil
}

func getFuncCallStringArg(funcCall *genai.FunctionCall, key, missingLabel, castLabel string) (string, error) {
	raw, ok := funcCall.Args[key]
	if !ok {
		return "", fmt.Errorf("there is no %v key inside func call args", missingLabel)
	}

	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast func call %v to string", castLabel)
	}

	return value, nil
}

type getFuncDeclarationArgs struct {
	userID string
}

type getFuncDeclarationResult struct {
	declarations []*genai.FunctionDeclaration
	toolTypes    map[string]commontype.ToolType
}

func (a *adapterImpl) getFuncDeclaration(ctx context.Context, args getFuncDeclarationArgs) (getFuncDeclarationResult, error) {
	tools, err := a.metadataUsecase.GetTool(ctx, metadata.GetToolArgs{
		UserID: args.userID,
	})
	if err != nil {
		return getFuncDeclarationResult{}, fmt.Errorf("unable to get tools: %w", err)
	}

	funcDeclarations := []*genai.FunctionDeclaration{}
	toolTypes := map[string]commontype.ToolType{}
	for _, r := range tools.Rows {
		declaration, tableLocation, err := buildFuncDeclaration(r)
		if err != nil {
			return getFuncDeclarationResult{}, err
		}

		funcDeclarations = append(funcDeclarations, declaration)
		toolTypes[tableLocation] = r.Type
	}

	return getFuncDeclarationResult{
		declarations: funcDeclarations,
		toolTypes:    toolTypes,
	}, nil
}

func buildFuncDeclaration(r metadata.GetToolRow) (*genai.FunctionDeclaration, string, error) {
	if !r.Type.ValidateToolType() {
		return nil, "", errors.New("invalid tool type")
	}

	tableLocation := buildTableLocation(r)

	return &genai.FunctionDeclaration{
		Name: tableLocation,
		Description: fmt.Sprintf(`
				Run select-only query from %s_hashed_filter to get
				information about: %s.

				%s

				The view has these columns:
				%s

				Sample query using the view:
				%s
				`, tableLocation,
			r.ToolDescription,
			buildBuilderQuery(r),
			buildColumnsDescription(r),
			buildExamples(r)),

		Parameters: buildParameterSchema(r),
		Response:   buildResponseSchema(),
	}, tableLocation, nil
}

func buildTableLocation(r metadata.GetToolRow) string {
	return fmt.Sprintf("%s.%s.%s",
		r.Project,
		r.Dataset,
		r.TableName)
}

func buildColumnsDescription(r metadata.GetToolRow) string {
	columns := []string{}
	for _, c := range r.Columns {
		columns = append(columns, fmt.Sprintf("- %v (%v): %v",
			c.Name, c.Type, c.Description))
	}

	return strings.Join(columns, "\n")
}

func buildBuilderQuery(r metadata.GetToolRow) string {
	if r.Type != commontype.QueryToolType {
		return ""
	}

	return fmt.Sprintf(`
			That view is build using this builder query:
			%v
			`, r.Examples[0].Query)
}

func buildExamples(r metadata.GetToolRow) string {
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

	return strings.Join(examples, "\n")
}

func buildParameterSchema(r metadata.GetToolRow) *genai.Schema {
	return &genai.Schema{
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
	}
}

func buildResponseSchema() *genai.Schema {
	return &genai.Schema{
		Type:        genai.TypeArray,
		Items:       &genai.Schema{Type: genai.TypeString},
		Description: "List of information returned by query, 1 item represents 1 row",
	}
}
