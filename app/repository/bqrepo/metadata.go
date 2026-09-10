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

func (r *repositoryImpl) CallTool(ctx context.Context, args CallToolArgs) (CallToolResult, error) {
	googleAuthClient := r.googleAuth.Client(ctx, googleauthadapter.ClientArgs{
		Token: args.Token,
	})

	bqClient, err := bigquery.NewClient(ctx, r.projectID, option.WithHTTPClient(googleAuthClient.Client))
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to create bigquery client: %w", err)
	}

	tableNameParts := strings.Split(args.TableName, ".")
	if len(tableNameParts) != 3 {
		return CallToolResult{}, fmt.Errorf("table name must be in the form of project.dataset.table")
	}

	defer bqClient.Close()
	defer func(ctx context.Context, args dropHashedFilterViewArgs) {
		dropErr := dropHashedFilterView(ctx, args)
		if dropErr != nil && err == nil {
			err = dropErr
		}
	}(ctx, dropHashedFilterViewArgs{
		bqClient:       bqClient,
		tableNameParts: tableNameParts,
	})

	err = createHashedFilterView(ctx, createHashedFilterViewArgs{
		bqClient:       bqClient,
		tableNameParts: tableNameParts,
	})
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to create hashed filter view: %w", err)
	}

	selectResult, err := selectHashedFilterView(ctx, selectHashedFilterViewArgs{
		bqClient:    bqClient,
		RawToolArgs: args.RawToolArgs,
	})
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to select hashed filter view: %w", err)
	}

	err = dropHashedFilterView(ctx, dropHashedFilterViewArgs{
		bqClient:       bqClient,
		tableNameParts: tableNameParts,
	})
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to drop hashed filter view: %w", err)
	}

	return CallToolResult{
		Rows: selectResult.rows,
	}, nil
}

type createHashedFilterViewArgs struct {
	bqClient       *bigquery.Client
	tableNameParts []string
	currentEpoch   int64
}

func createHashedFilterView(ctx context.Context, args createHashedFilterViewArgs) error {
	filter, ok := ctx.Value(commonctx.FilterCtxKey).(string)
	if !ok {
		return fmt.Errorf("unable to get filter from context")
	}

	salt, ok := ctx.Value(commonctx.SaltCtxKey).(string)
	if !ok {
		return fmt.Errorf("unable to get salt from context")
	}

	err := args.bqClient.Dataset(args.tableNameParts[1]).
		Table(args.tableNameParts[2]+"_hashed_filter").
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
				args.tableNameParts[1]+"."+args.tableNameParts[2]),
		})
	if err != nil {
		return fmt.Errorf("unable to create view containing hashed filter: %v", err)
	}

	return nil
}

type selectHashedFilterViewArgs struct {
	bqClient    *bigquery.Client
	RawToolArgs []byte
}

type hashedFilterViewToolArgs struct {
	Query string `json:"query"`
}

type selectHashedFilterViewResult struct {
	rows []map[string]bigquery.Value
}

func selectHashedFilterView(ctx context.Context, args selectHashedFilterViewArgs) (selectHashedFilterViewResult, error) {
	var toolArgs hashedFilterViewToolArgs
	err := json.Unmarshal(args.RawToolArgs, &toolArgs)
	if err != nil {
		return selectHashedFilterViewResult{}, fmt.Errorf("unable to unmarshal tool args: %w", err)
	}

	selectJob, err := args.bqClient.Query(toolArgs.Query).
		Run(ctx)
	if err != nil {
		return selectHashedFilterViewResult{}, fmt.Errorf("unable to run select job from hashed filter view: %w", err)
	}

	selectJobStatus, err := selectJob.Wait(ctx)
	if err != nil {
		return selectHashedFilterViewResult{}, fmt.Errorf("select job has failed: %w", err)
	}

	if selectJobStatus.Err() != nil {
		return selectHashedFilterViewResult{}, fmt.Errorf("select job has error: %w", selectJobStatus.Err())
	}

	selectJobRows, err := selectJob.Read(ctx)
	if err != nil {
		return selectHashedFilterViewResult{}, fmt.Errorf("unable to get result of select job: %w", err)
	}

	rows := []map[string]bigquery.Value{}
	for {
		var row map[string]bigquery.Value
		err := selectJobRows.Next(&row)
		if err == iterator.Done {
			break
		}

		if err != nil {
			return selectHashedFilterViewResult{}, fmt.Errorf("row doesn't conform to resultRow map: %w", err)
		}

		rows = append(rows, row)
	}

	return selectHashedFilterViewResult{
		rows: rows,
	}, nil
}

type dropHashedFilterViewArgs struct {
	bqClient       *bigquery.Client
	tableNameParts []string
}

func dropHashedFilterView(ctx context.Context, args dropHashedFilterViewArgs) error {
	err := args.bqClient.Dataset(args.tableNameParts[1]).
		Table(args.tableNameParts[2] + "_hashed_filter").
		Delete(ctx)
	if err != nil {
		return fmt.Errorf("unable to drop hashed filter view: %w", err)
	}

	return nil

}
