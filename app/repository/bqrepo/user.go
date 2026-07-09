package bqrepo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/gatsu420/kisu-be/common/commonhash"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Aggregation struct {
	Func   string
	Column string
}

type GetUserArgs struct {
	Token        *oauth2.Token
	Columns      []string
	Emails       []string
	Aggregations []Aggregation
}

func (r *repositoryImpl) GetUser(ctx context.Context, args GetUserArgs) ([]map[string]bigquery.Value, error) {
	var groupByColumns []string
	if args.Aggregations != nil {
		for _, c := range args.Columns {
			var isGroupByColumn bool

			for _, a := range args.Aggregations {
				if c == a.Column {
					isGroupByColumn = true
					break
				}
			}

			if !isGroupByColumn {
				groupByColumns = append(groupByColumns, c)
			}
		}
	}

	var selectedColumns []string
	for _, c := range groupByColumns {
		selectedColumns = append(selectedColumns, c)
	}
	for _, a := range args.Aggregations {
		selectedColumns = append(selectedColumns,
			fmt.Sprintf("%v(%v) as %v_%v", a.Func, a.Column, a.Func, a.Column))
	}

	googleAuthClient := r.googleAuth.Client(ctx, args.Token)
	bqClient, err := bigquery.NewClient(ctx, r.projectID, option.WithHTTPClient(googleAuthClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create bigquery client: %w", err)
	}

	query := fmt.Sprintf(`
		select %v from (
			select
				'test@gmail.com' email,
				'test user' name,
				22 age

			union all

			select
				'anothertest@gmail.com' email,
				'another test user' name,
				35 age

			union all

			select
				'fakeuser@hotmail.com' email,
				'fake user' name,
				30 age
		)
		where to_base64(sha256(concat(email, @salt))) in unnest(@emails)
		`, strings.Join(selectedColumns, ","))
	for i := range groupByColumns {
		if i == 0 {
			query += ("\ngroup by " + strconv.Itoa(i+1))
		} else {
			query += ("," + strconv.Itoa(i+1))
		}
	}
	bqQuery := bqClient.Query(query)

	randomIntSaltPart, ok := ctx.Value(commonhash.RandomIntCtxKey).(int)
	if !ok {
		return nil, fmt.Errorf("unable to get random int salt part from context")
	}

	salt := r.stringSaltPart + strconv.Itoa(randomIntSaltPart)
	bqQuery.Parameters = []bigquery.QueryParameter{
		{Name: "emails", Value: args.Emails},
		{Name: "salt", Value: salt},
	}
	rows, err := bqQuery.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	result := []map[string]bigquery.Value{}
	for {
		var item map[string]bigquery.Value
		err := rows.Next(&item)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("item doesn't conform to result: %w", err)
		}

		result = append(result, item)
	}

	return result, nil
}
