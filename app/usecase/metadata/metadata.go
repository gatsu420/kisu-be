package metadata

import (
	"context"
	"fmt"

	"github.com/gatsu420/kisu-be/app/repository/pgrepo"
	"golang.org/x/oauth2"
)

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
	UserID           string
	ToolDescription  string
	TableName        string
	Columns          []AddToolColumn
	QueryExamples    []AddToolQueryExample
	ParamName        string
	ParamType        string
	ParamDescription string
}

type AddToolColumn struct {
	Name        string
	Type        string
	Description string
}

type AddToolQueryExample struct {
	Description string
	Query       string
}

func (u *usecaseImpl) AddTool(ctx context.Context, args AddToolArgs) error {
	columns := []pgrepo.AddToolColumn{}
	for _, c := range args.Columns {
		columns = append(columns, pgrepo.AddToolColumn{
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
		})
	}

	queryExamples := []pgrepo.AddToolQueryExample{}
	for _, qe := range args.QueryExamples {
		queryExamples = append(queryExamples, pgrepo.AddToolQueryExample{
			Description: qe.Description,
			Query:       qe.Query,
		})
	}

	err := u.pgRepo.AddTool(ctx, pgrepo.AddToolArgs{
		UserID:           args.UserID,
		ToolDescription:  args.ToolDescription,
		TableName:        args.TableName,
		Columns:          columns,
		QueryExamples:    queryExamples,
		ParamName:        args.ParamName,
		ParamType:        args.ParamType,
		ParamDescription: args.ParamDescription,
	})
	if err != nil {
		return fmt.Errorf("unable to add tool: %w", err)
	}

	return nil
}

type GetToolArgs struct {
	UserID string
}

type GetToolRow struct {
	ToolDescription  string
	TableName        string
	Columns          []GetToolColumn
	QueryExample     []GetToolQueryExample
	ParamName        string
	ParamType        string
	ParamDescription string
}

type GetToolColumn struct {
	Name        string
	Type        string
	Description string
}

type GetToolQueryExample struct {
	Description string
	Query       string
}

func (u *usecaseImpl) GetTool(ctx context.Context, args GetToolArgs) ([]GetToolRow, error) {
	rows, err := u.pgRepo.GetTool(ctx, pgrepo.GetToolArgs{
		UserID: args.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get tool: %w", err)
	}

	resultRows := []GetToolRow{}
	for _, r := range rows {
		resultColumns := []GetToolColumn{}
		for _, c := range r.Columns {
			resultColumns = append(resultColumns, GetToolColumn{
				Name:        c.Name,
				Type:        c.Type,
				Description: c.Description,
			})
		}

		resultQueryExamples := []GetToolQueryExample{}
		for _, qe := range r.QueryExample {
			resultQueryExamples = append(resultQueryExamples, GetToolQueryExample{
				Description: qe.Description,
				Query:       qe.Query,
			})
		}

		resultRows = append(resultRows, GetToolRow{
			ToolDescription:  r.ToolDescription,
			TableName:        r.TableName,
			Columns:          resultColumns,
			QueryExample:     resultQueryExamples,
			ParamName:        r.ParamName,
			ParamType:        r.ParamType,
			ParamDescription: r.ParamDescription,
		})
	}

	return resultRows, nil
}
