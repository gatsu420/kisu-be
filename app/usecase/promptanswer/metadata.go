package promptanswer

import (
	"context"
	"fmt"

	"github.com/gatsu420/kisu-be/app/repository/pgrepo"
)

type Column struct {
	Name        string
	Type        string
	Description string
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

func (u *usecaseImpl) AddMetadata(ctx context.Context, args AddMetadataArgs) error {
	columns := []pgrepo.Column{}
	for _, c := range args.Columns {
		columns = append(columns, pgrepo.Column{
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
		})
	}

	queryExamples := []pgrepo.QueryExample{}
	for _, e := range args.QueryExamples {
		queryExamples = append(queryExamples, pgrepo.QueryExample{
			Description: e.Description,
			Query:       e.Query,
		})
	}

	err := u.pgRepo.AddMetadata(ctx, pgrepo.AddMetadataArgs{
		ToolDescription: args.ToolDescription,
		TableName:       args.TableName,
		Columns:         columns,
		QueryExamples:   queryExamples,
	})
	if err != nil {
		return fmt.Errorf("unable to add metadata: %w", err)
	}

	return nil
}
