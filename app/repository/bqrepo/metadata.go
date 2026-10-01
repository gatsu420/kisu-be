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

	job, err := bqClient.Query(queryPrefix + args.Query).
		Run(ctx)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to run select job from hashed filter view: %w", err)
	}

	jobStatus, err := job.Wait(ctx)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("select job has failed: %w", err)
	}

	if jobStatus.Err() != nil {
		return CallToolResult{}, fmt.Errorf("select job has error: %w", jobStatus.Err())
	}

	resultRows, err := job.Read(ctx)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to get result of select job: %w", err)
	}

	rows := []map[string]bigquery.Value{}
	for {
		var row map[string]bigquery.Value
		err := resultRows.Next(&row)
		if err == iterator.Done {
			break
		}

		if err != nil {
			return CallToolResult{}, fmt.Errorf("row doesn't conform to row map: %w", err)
		}

		rows = append(rows, row)
	}

	return CallToolResult{
		Rows: rows,
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
