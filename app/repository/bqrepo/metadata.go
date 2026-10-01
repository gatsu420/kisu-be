package bqrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/common/commontype"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type CallToolArgs struct {
	Filter        string
	Salt          string
	Type          commontype.ToolType
	TableLocation string
	BuilderQuery  string
	Query         string
	Token         *oauth2.Token
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

	tableLocationParts := strings.Split(args.TableLocation, ".")
	if len(tableLocationParts) != 3 {
		return CallToolResult{}, errors.New("table location must be in the form of project.dataset.table")
	}

	result, err := runQuery(ctx, runQueryArgs{
		bqClient:     bqClient,
		filter:       args.Filter,
		salt:         args.Salt,
		toolType:     args.Type,
		project:      tableLocationParts[0],
		dataset:      tableLocationParts[1],
		tableName:    tableLocationParts[2],
		builderQuery: args.BuilderQuery,
		query:        args.Query,
	})
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to select CTE: %w", err)
	}

	return CallToolResult{
		Rows: result.rows,
	}, nil
}

type runQueryArgs struct {
	bqClient     *bigquery.Client
	filter       string
	salt         string
	toolType     commontype.ToolType
	project      string
	dataset      string
	tableName    string
	builderQuery string
	query        string
}

type runQueryResult struct {
	rows []map[string]bigquery.Value
}

func runQuery(ctx context.Context, args runQueryArgs) (runQueryResult, error) {
	queryPrefix := fmt.Sprintf(`
	with hashed_filter as (
		select
			*,
			to_base64(sha256(concat(%s, "%s"))) hashed_%s
		from
	`, args.filter, args.salt, args.filter)

	switch args.toolType {
	case commontype.TableToolType:
		queryPrefix += fmt.Sprintf(" %s.%s.%s\n)\n",
			args.project,
			args.dataset,
			args.tableName)
	case commontype.QueryToolType:
		queryPrefix += fmt.Sprintf(" (%s)\n)\n",
			args.builderQuery)
	}

	job, err := args.bqClient.Query(queryPrefix + args.query).
		Run(ctx)
	if err != nil {
		return runQueryResult{}, fmt.Errorf("unable to run select job from hashed filter view: %w", err)
	}

	jobStatus, err := job.Wait(ctx)
	if err != nil {
		return runQueryResult{}, fmt.Errorf("select job has failed: %w", err)
	}

	if jobStatus.Err() != nil {
		return runQueryResult{}, fmt.Errorf("select job has error: %w", jobStatus.Err())
	}

	resultRows, err := job.Read(ctx)
	if err != nil {
		return runQueryResult{}, fmt.Errorf("unable to get result of select job: %w", err)
	}

	rows := []map[string]bigquery.Value{}
	for {
		var row map[string]bigquery.Value
		err := resultRows.Next(&row)
		if err == iterator.Done {
			break
		}

		if err != nil {
			return runQueryResult{}, fmt.Errorf("row doesn't conform to row map: %w", err)
		}

		rows = append(rows, row)
	}

	return runQueryResult{
		rows: rows,
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
