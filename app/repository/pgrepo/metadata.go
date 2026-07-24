package pgrepo

import (
	"context"
	"fmt"
)

type Column struct {
	Name         string
	Type         string
	Description  string
	IsPersistent bool
}

type QueryExample struct {
	Description string
	Query       string
}

type AddMetadataArgs struct {
	ToolDescription string
	TableName       string
	Columns         []Column
	QueryExamples   []QueryExample
}

func (r *repositoryImpl) AddMetadata(ctx context.Context, args AddMetadataArgs) error {
	_, err := r.pgxPool.Exec(ctx, `
	insert into metadata (
		tool_description, table_name, columns, query_examples
	) values ($1, $2, $3, $4, $5, $6)
	`, args.ToolDescription, args.TableName, args.Columns, args.QueryExamples)
	if err != nil {
		return fmt.Errorf("unable to execute query: %w", err)
	}

	return nil
}
