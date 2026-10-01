package bqrepo

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/common/commontype"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type CallToolArgs struct {
	Filter       string
	Salt         string
	Type         commontype.ToolType
	Project      string
	Dataset      string
	TableName    string
	BuilderQuery string
	Query        string
	Token        *oauth2.Token
}

type CallToolResult struct {
	Rows []map[string]bigquery.Value
}

func (r *repositoryImpl) CallTool(ctx context.Context, args CallToolArgs) (CallToolResult, error) {
	googleAuthClient := r.googleAuth.Client(ctx, googleauthadapter.ClientArgs{
		Token: args.Token,
	})

	bqClient, err := bigquery.NewClient(ctx, r.projectID, option.WithHTTPClient(googleAuthClient.Client))
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to create bigquery client: %w", err)
	}
	defer bqClient.Close()

	queryPrefix := fmt.Sprintf(`
	with hashed_filter as (
		select
			*,
			to_base64(sha256(concat(%s, "%s"))) hashed_%s
		from
	`, args.Filter, args.Salt, args.Filter)

	switch args.Type {
	case commontype.TableToolType:
		queryPrefix += fmt.Sprintf(" %s.%s.%s\n)\n",
			args.Project,
			args.Dataset,
			args.TableName)
	case commontype.QueryToolType:
		queryPrefix += fmt.Sprintf(" (%s)\n)\n",
			args.BuilderQuery)
	}

	result, err := runQueryWithRows(ctx, runQueryWithRowsArgs{
		bqClient: bqClient,
		query:    queryPrefix + args.Query,
		token:    args.Token,
	})
	if err != nil {
		return CallToolResult{}, err
	}

	return CallToolResult{
		Rows: result.rows,
	}, nil
}

type ValidateToolQueryArgs struct {
	Query string
	Token *oauth2.Token
}

type ValidateToolQueryResult struct {
	IsValid bool
}

func (r *repositoryImpl) ValidateToolQuery(ctx context.Context, args ValidateToolQueryArgs) (ValidateToolQueryResult, error) {
	googleAuthClient := r.googleAuth.Client(ctx,
		googleauthadapter.ClientArgs{
			Token: args.Token,
		})

	bqClient, err := bigquery.NewClient(ctx,
		r.projectID,
		option.WithHTTPClient(googleAuthClient.Client))
	if err != nil {
		return ValidateToolQueryResult{}, fmt.Errorf("unable to create bigquery client: %w", err)
	}
	defer bqClient.Close()

	query := bqClient.Query(args.Query)
	query.DryRun = true
	job, err := query.Run(ctx)
	if err != nil {
		return ValidateToolQueryResult{}, fmt.Errorf("unable to run job for validating query: %w", err)
	}

	if job.LastStatus().Err() != nil {
		return ValidateToolQueryResult{}, fmt.Errorf("job for validating query has failed: %w", err)
	}

	return ValidateToolQueryResult{
		IsValid: true,
	}, nil
}

type GetToolTableMetadataArgs struct {
	Type         commontype.ToolType
	Project      string
	Dataset      string
	TableName    string
	BuilderQuery string
	Token        *oauth2.Token
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

func (r *repositoryImpl) GetToolTableMetadata(ctx context.Context, args GetToolTableMetadataArgs) (GetToolTableMetadataResult, error) {
	googleAuthClient := r.googleAuth.Client(ctx,
		googleauthadapter.ClientArgs{
			Token: args.Token,
		})

	bqClient, err := bigquery.NewClient(ctx,
		r.projectID,
		option.WithHTTPClient(googleAuthClient.Client))
	if err != nil {
		return GetToolTableMetadataResult{}, fmt.Errorf("unable to create bigquery client: %w", err)
	}
	defer bqClient.Close()

	metadata, err := bqClient.DatasetInProject(args.Project, args.Dataset).
		Table(args.TableName).
		Metadata(ctx)
	if err != nil {
		return GetToolTableMetadataResult{}, fmt.Errorf("unable to get metadata: %w", err)
	}

	columns := []GetToolTableMetadataColumn{}
	switch args.Type {
	case commontype.TableToolType:
		metadata, err := bqClient.DatasetInProject(args.Project, args.Dataset).
			Table(args.TableName).
			Metadata(ctx)
		if err != nil {
			return GetToolTableMetadataResult{}, fmt.Errorf("unable to get metadata: %w", err)
		}

		for _, s := range metadata.Schema {
			columns = append(columns, GetToolTableMetadataColumn{
				Name:        s.Name,
				Type:        string(s.Type),
				Description: s.Description,
			})
		}

	case commontype.QueryToolType:
		result, err := runQueryWithRows(ctx, runQueryWithRowsArgs{
			bqClient: bqClient,
			query:    args.BuilderQuery,
			token:    args.Token,
		})
		if err != nil {
			return GetToolTableMetadataResult{}, err
		}

		for _, s := range result.schema {
			columns = append(columns, GetToolTableMetadataColumn{
				Name:        s.name,
				Type:        s.dataType,
				Description: s.description,
			})
		}
	}

	return GetToolTableMetadataResult{
		Description: metadata.Description,
		Columns:     columns,
	}, nil
}

type runQueryWithRowsArgs struct {
	bqClient *bigquery.Client
	query    string
	token    *oauth2.Token
}

type runQueryWithRowsResult struct {
	rows   []map[string]bigquery.Value
	schema []runQueryWithRowsSchema
}

type runQueryWithRowsSchema struct {
	name        string
	dataType    string
	description string
}

func runQueryWithRows(ctx context.Context, args runQueryWithRowsArgs) (runQueryWithRowsResult, error) {
	job, err := args.bqClient.Query(args.query).
		Run(ctx)
	if err != nil {
		return runQueryWithRowsResult{}, fmt.Errorf("unable to run select job from hashed filter view: %w", err)
	}

	jobStatus, err := job.Wait(ctx)
	if err != nil {
		return runQueryWithRowsResult{}, fmt.Errorf("select job has failed: %w", err)
	}

	if jobStatus.Err() != nil {
		return runQueryWithRowsResult{}, fmt.Errorf("select job has error: %w", jobStatus.Err())
	}

	resultRows, err := job.Read(ctx)
	if err != nil {
		return runQueryWithRowsResult{}, fmt.Errorf("unable to get result of select job: %w", err)
	}

	rows := []map[string]bigquery.Value{}
	for {
		var row map[string]bigquery.Value
		err := resultRows.Next(&row)
		if err == iterator.Done {
			break
		}

		if err != nil {
			return runQueryWithRowsResult{}, fmt.Errorf("row doesn't conform to row map: %w", err)
		}

		rows = append(rows, row)
	}

	schema := []runQueryWithRowsSchema{}
	for _, s := range resultRows.Schema {
		schema = append(schema, runQueryWithRowsSchema{
			name:        s.Name,
			dataType:    string(s.Type),
			description: s.Description,
		})
	}

	return runQueryWithRowsResult{
		rows:   rows,
		schema: schema,
	}, nil
}
