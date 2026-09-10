package bqrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/common/commonctx"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type CallToolArgs struct {
	TableName   string
	RawToolArgs []byte
	Token       *oauth2.Token
}

type CallToolResult struct {
	Rows []map[string]bigquery.Value
}

type toolArgs struct {
	Query string `json:"query"`
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

	var toolArgs toolArgs
	err = json.Unmarshal(args.RawToolArgs, &toolArgs)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to unmarshal tool args: %w", err)
	}

	filter, ok := ctx.Value(commonctx.FilterCtxKey).(string)
	if !ok {
		return CallToolResult{}, fmt.Errorf("unable to get filter from context")
	}

	salt, ok := ctx.Value(commonctx.SaltCtxKey).(string)
	if !ok {
		return CallToolResult{}, fmt.Errorf("unable to get salt from context")
	}

	tableNameParts := strings.Split(args.TableName, ".")
	if len(tableNameParts) != 3 {
		return CallToolResult{}, fmt.Errorf("table name must be in the form of project.dataset.table")
	}

	err = bqClient.Dataset(tableNameParts[1]).
		Table(tableNameParts[2]+"_hashed_filter").
		Create(ctx, &bigquery.TableMetadata{
			ViewQuery: fmt.Sprintf(`
			select
				* except(%v),
				to_base64(sha256(concat(%v, "%v"))) hashed_%v
			from %v
			`, filter,
				filter,
				salt,
				filter,
				args.TableName),
		})
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to create view containing hashed filter: %v", err)
	}

	selectJob, err := bqClient.Query(toolArgs.Query).
		Run(ctx)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to run select job from hashed filter view: %w", err)
	}

	selectJobStatus, err := selectJob.Wait(ctx)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("select job has failed: %w", err)
	}

	if selectJobStatus.Err() != nil {
		return CallToolResult{}, fmt.Errorf("select job has error: %w", selectJobStatus.Err())
	}

	selectJobRows, err := selectJob.Read(ctx)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to get result of select job: %w", err)
	}

	resultRows := []map[string]bigquery.Value{}
	for {
		var resultRow map[string]bigquery.Value
		err := selectJobRows.Next(&resultRow)
		if err == iterator.Done {
			break
		}

		if err != nil {
			return CallToolResult{}, fmt.Errorf("row doesn't conform to resultRow map: %w", err)
		}

		resultRows = append(resultRows, resultRow)
	}

	err = bqClient.Dataset(tableNameParts[1]).
		Table(tableNameParts[2] + "_hashed_filter").
		Delete(ctx)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to drop hashed filter view: %w", err)
	}

	return CallToolResult{
		Rows: resultRows,
	}, nil

}
