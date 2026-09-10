package bqrepo

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/common/commonhash"
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

	var toolArgs toolArgs
	err = json.Unmarshal(args.RawToolArgs, &toolArgs)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to unmarshal tool args: %w", err)
	}

	filter, ok := ctx.Value(commonhash.FilterCtxKey).(string)
	if !ok {
		return CallToolResult{}, fmt.Errorf("unable to get filter from context")
	}

	salt, ok := ctx.Value(commonhash.SaltCtxKey).(string)
	if !ok {
		return CallToolResult{}, fmt.Errorf("unable to get salt from context")
	}

	hashQuery := bqClient.Query(fmt.Sprintf(`
		create view %v_hashed_filter as
		select
			* except(%v),
			to_base64(sha256(concat(%v, "%v"))) hashed_%v
		from %v
		`, args.TableName,
		filter,
		filter,
		salt,
		filter,
		args.TableName))
	_, err = hashQuery.Run(ctx)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to run job for hash query: %w", err)
	}

	getterQuery := bqClient.Query(toolArgs.Query)
	rows, err := getterQuery.Read(ctx)
	if err != nil {
		return CallToolResult{}, fmt.Errorf("unable to run job for getter query: %w", err)
	}

	resultRows := []map[string]bigquery.Value{}
	for {
		var resultRow map[string]bigquery.Value
		err := rows.Next(&resultRow)
		if err == iterator.Done {
			break
		}

		if err != nil {
			return CallToolResult{}, fmt.Errorf("row doesn't conform to resultRow map: %w", err)
		}

		resultRows = append(resultRows, resultRow)
	}

	return CallToolResult{
		Rows: resultRows,
	}, nil

}
