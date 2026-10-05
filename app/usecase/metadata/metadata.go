package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gatsu420/kisu-be/app/repository/bqrepo"
	"github.com/gatsu420/kisu-be/app/repository/pgrepo"
	"github.com/gatsu420/kisu-be/common/commoncrypto"
	"github.com/gatsu420/kisu-be/common/commonctx"
	"github.com/gatsu420/kisu-be/common/commontype"
	"golang.org/x/oauth2"
	"google.golang.org/genai"
)

type AddAuthStateArgs struct {
	State string
}

func (u *usecaseImpl) AddAuthState(ctx context.Context, args AddAuthStateArgs) error {
	return u.pgRepo.AddAuthState(ctx, pgrepo.AddAuthStateArgs{
		State: args.State,
	})
}

type ConsumeAuthStateArgs struct {
	State string
}

type ConsumeAuthStateResult struct {
	StateExistence bool
}

func (u *usecaseImpl) ConsumeAuthState(ctx context.Context, args ConsumeAuthStateArgs) (ConsumeAuthStateResult, error) {
	result, err := u.pgRepo.ConsumeAuthState(ctx, pgrepo.ConsumeAuthStateArgs{
		State: args.State,
	})
	if err != nil {
		return ConsumeAuthStateResult{}, err
	}

	return ConsumeAuthStateResult{
		StateExistence: result.StateExistence,
	}, nil
}

type AddUserArgs struct {
	Email string
}

type AddUserResult struct {
	UserID string
}

func (u *usecaseImpl) AddUser(ctx context.Context, args AddUserArgs) (AddUserResult, error) {
	result, err := u.pgRepo.AddUser(ctx, pgrepo.AddUserArgs{
		Email: args.Email,
	})
	if err != nil {
		return AddUserResult{}, err
	}

	return AddUserResult{
		UserID: result.UserID,
	}, nil
}

type AddUserTokenArgs struct {
	UserID string
	Token  *oauth2.Token
}

func (u *usecaseImpl) AddUserToken(ctx context.Context, args AddUserTokenArgs) error {
	err := u.pgRepo.AddUserToken(ctx, pgrepo.AddUserTokenArgs{
		UserID: args.UserID,
		Token:  args.Token,
	})
	if err != nil {
		return err
	}

	return nil
}

type GetUserTokenArgs struct {
	UserID string
}

type GetUserTokenResult struct {
	Token *oauth2.Token
}

func (u *usecaseImpl) GetUserToken(ctx context.Context, args GetUserTokenArgs) (GetUserTokenResult, error) {
	result, err := u.pgRepo.GetUserToken(ctx, pgrepo.GetUserTokenArgs{
		UserID: args.UserID,
	})
	if err != nil {
		return GetUserTokenResult{}, err
	}

	return GetUserTokenResult{
		Token: result.Token,
	}, nil
}

type AddToolArgs struct {
	UserID          string
	ToolDescription string
	Project         string
	Dataset         string
	TableName       string
	Columns         []AddToolColumn
	Type            commontype.ToolType
	Examples        []AddToolExample
	ParamNames      []string
}

type AddToolColumn struct {
	Name        string
	Type        string
	Description string
}

type AddToolExample struct {
	Description string
	Query       string
}

func (u *usecaseImpl) AddTool(ctx context.Context, args AddToolArgs) error {
	if !args.Type.ValidateToolType() {
		return errors.New("invalid type value")
	}

	if args.Type == commontype.TableToolType {
		if args.Project == "" || args.Dataset == "" || args.TableName == "" {
			return errors.New("table tool requires all location components to be supplied")
		}
	}

	if args.Type == commontype.QueryToolType {
		if len(args.Examples) > 1 {
			return errors.New("query tool requires just one example")
		}

		if args.Examples[0].Description != "" {
			return errors.New("query tool requires example description to not be supplied")
		}

		args.TableName = commoncrypto.GetRandomTableName()
	}

	columns := []pgrepo.AddToolColumn{}
	for _, c := range args.Columns {
		columns = append(columns, pgrepo.AddToolColumn{
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
		})
	}

	examples := []pgrepo.AddToolExample{}
	for _, e := range args.Examples {
		examples = append(examples, pgrepo.AddToolExample{
			Description: e.Description,
			Query:       e.Query,
		})
	}

	return u.pgRepo.AddTool(ctx, pgrepo.AddToolArgs{
		UserID:          args.UserID,
		ToolDescription: args.ToolDescription,
		Project:         args.Project,
		Dataset:         args.Dataset,
		TableName:       args.TableName,
		Columns:         columns,
		Type:            args.Type,
		Examples:        examples,
		ParamNames:      args.ParamNames,
	})
}

type GetToolArgs struct {
	UserID string
}

type GetToolResult struct {
	Rows []GetToolRow
}

type GetToolRow struct {
	ID              string
	ToolDescription string
	Project         string
	Dataset         string
	TableName       string
	Columns         []GetToolColumn
	Type            commontype.ToolType
	Examples        []GetToolExample
	ParamNames      []string
}

type GetToolColumn struct {
	Name        string
	Type        string
	Description string
}

type GetToolExample struct {
	Description string
	Query       string
}

func (u *usecaseImpl) GetTool(ctx context.Context, args GetToolArgs) (GetToolResult, error) {
	rows, err := u.pgRepo.GetTool(ctx, pgrepo.GetToolArgs{
		UserID: args.UserID,
	})
	if err != nil {
		return GetToolResult{}, err
	}

	resultRows := []GetToolRow{}
	for _, r := range rows.Rows {
		var dataset string
		if r.Dataset != nil {
			dataset = *r.Dataset
		}

		var tableName string
		if r.TableName != nil {
			tableName = *r.TableName
		}

		resultColumns := []GetToolColumn{}
		for _, c := range r.Columns {
			resultColumns = append(resultColumns, GetToolColumn{
				Name:        c.Name,
				Type:        c.Type,
				Description: c.Description,
			})
		}

		resultExamples := []GetToolExample{}
		for _, e := range r.Examples {
			resultExamples = append(resultExamples, GetToolExample{
				Description: e.Description,
				Query:       e.Query,
			})
		}

		resultRows = append(resultRows, GetToolRow{
			ID:              r.ID,
			ToolDescription: r.ToolDescription,
			Project:         r.Project,
			Dataset:         dataset,
			TableName:       tableName,
			Columns:         resultColumns,
			Type:            r.Type,
			Examples:        resultExamples,
			ParamNames:      r.ParamNames,
		})
	}

	return GetToolResult{
		Rows: resultRows,
	}, nil
}

type CallToolArgs struct {
	Tool   json.RawMessage
	Type   commontype.ToolType
	Limit  string
	Offset string
}

type CallToolResult struct {
	Result          json.RawMessage
	StringifiedTool string
}

func (u *usecaseImpl) CallTool(ctx context.Context, args CallToolArgs) (CallToolResult, error) {
	token, ok := ctx.Value(commonctx.TokenCtxKey).(*oauth2.Token)
	if !ok {
		return CallToolResult{}, errors.New("unable to get token from context")
	}

	paramName, ok := ctx.Value(commonctx.ParamNameCtxKey).(string)
	if !ok {
		return CallToolResult{}, errors.New("unable to get param name from context")
	}

	salt, ok := ctx.Value(commonctx.SaltCtxKey).(string)
	if !ok {
		return CallToolResult{}, errors.New("unable to get salt from context")
	}

	var tool *genai.FunctionCall
	err := json.Unmarshal(args.Tool, &tool)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to unmarshal tool: %w", err)
	}

	var project, dataset, tableName string
	if args.Type == commontype.TableToolType {
		nameParts := strings.Split(tool.Name, ".")
		if len(nameParts) != 3 {
			return CallToolResult{}, errors.New("name for table tool must be in the form of \"{string}.{string}.{string}\"")
		}

		project, dataset, tableName = nameParts[0], nameParts[1], nameParts[2]
	}

	builderQuery, ok := tool.Args["builder_query"]
	if !ok {
		return CallToolResult{}, errors.New("there is no builder_query key inside tool args")
	}

	stringifiedBuilderQuery, ok := builderQuery.(string)
	if !ok {
		return CallToolResult{}, errors.New("unable to cast tool builder query to string")
	}

	query, ok := tool.Args["query"]
	if !ok {
		return CallToolResult{}, errors.New("there is no query key inside tool args")
	}

	stringifiedQuery, ok := query.(string)
	if !ok {
		return CallToolResult{}, errors.New("unable to cast tool query to string")
	}

	paramKey, ok := tool.Args["hashed_param"]
	if !ok {
		return CallToolResult{}, errors.New("there is no hashed param key inside tool args")
	}

	param, ok := paramKey.(map[string]any)
	if !ok {
		return CallToolResult{}, errors.New("unable to cast tool param key to map")
	}

	paramVal, ok := param[paramName]
	if !ok {
		return CallToolResult{}, fmt.Errorf("there is no %s key inside tool param", paramName)
	}

	stringifiedQuery += fmt.Sprintf(`
		where %s in (%s)
		order by %s limit %s offset %s
	`,
		"hashed_"+paramName, paramVal,
		paramName, args.Limit, args.Offset)

	marshaledTool, err := json.MarshalIndent(args.Tool, "", " ")
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to marshal tool: %w", err)
	}

	result, err := u.bqRepo.CallTool(ctx, bqrepo.CallToolArgs{
		ParamName:    paramName,
		Salt:         salt,
		Type:         args.Type,
		Project:      project,
		Dataset:      dataset,
		TableName:    tableName,
		BuilderQuery: stringifiedBuilderQuery,
		Query:        stringifiedQuery,
		Token:        token,
	})
	if err != nil {
		return CallToolResult{}, err
	}

	marshaledRows, err := json.Marshal(result.Rows)
	if err != nil {
		return CallToolResult{}, err
	}

	return CallToolResult{
		Result:          marshaledRows,
		StringifiedTool: string(marshaledTool),
	}, nil
}

type ValidateToolQueryArgs struct {
	Query string
}

type ValidateToolQueryResult struct {
	IsValid bool
}

func (u *usecaseImpl) ValidateToolQuery(ctx context.Context, args ValidateToolQueryArgs) (ValidateToolQueryResult, error) {
	token, ok := ctx.Value(commonctx.TokenCtxKey).(*oauth2.Token)
	if !ok {
		return ValidateToolQueryResult{}, errors.New("unable to get token from context")
	}

	result, err := u.bqRepo.ValidateToolQuery(ctx, bqrepo.ValidateToolQueryArgs{
		Query: args.Query,
		Token: token,
	})
	if err != nil {
		return ValidateToolQueryResult{}, err
	}

	return ValidateToolQueryResult{
		IsValid: result.IsValid,
	}, nil
}

type GetToolTableMetadataArgs struct {
	Type         commontype.ToolType
	Project      string
	Dataset      string
	TableName    string
	BuilderQuery string
}

type GetToolTableMetadataResult struct {
	Description string
	Columns     []GetToolTableMetadataColumn
}

type GetToolTableMetadataColumn struct {
	Name        string
	Type        string
	Description string
}

func (u *usecaseImpl) GetToolTableMetadata(ctx context.Context, args GetToolTableMetadataArgs) (GetToolTableMetadataResult, error) {
	token, ok := ctx.Value(commonctx.TokenCtxKey).(*oauth2.Token)
	if !ok {
		return GetToolTableMetadataResult{}, errors.New("unable to get token from context")
	}

	result, err := u.bqRepo.GetToolTableMetadata(ctx, bqrepo.GetToolTableMetadataArgs{
		Type:         args.Type,
		Project:      args.Project,
		Dataset:      args.Dataset,
		TableName:    args.TableName,
		BuilderQuery: args.BuilderQuery,
		Token:        token,
	})
	if err != nil {
		return GetToolTableMetadataResult{}, err
	}

	columns := []GetToolTableMetadataColumn{}
	for _, c := range result.Columns {
		columns = append(columns, GetToolTableMetadataColumn{
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
		})
	}

	return GetToolTableMetadataResult{
		Description: result.Description,
		Columns:     columns,
	}, nil
}
