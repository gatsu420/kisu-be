package bqrepo

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/gatsu420/kisu-be/common/commonhash"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GetUserArgs struct {
	Token   *oauth2.Token
	Columns []string
	Emails  []string
}

type GetUserItem struct {
	Email string `bigquery:"email"`
	Name  string `bigquery:"name"`
	Age   int    `bigquery:"age"`
}

func (r *repositoryImpl) GetUser(ctx context.Context, args GetUserArgs) ([]GetUserItem, error) {
	if !slices.Contains(args.Columns, "email") {
		args.Columns = append(args.Columns, "email")
	}
	selectedColumns := strings.Join(args.Columns, ",")

	googleAuthClient := r.googleAuth.Client(ctx, args.Token)
	bqClient, err := bigquery.NewClient(ctx, r.projectID, option.WithHTTPClient(googleAuthClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create bigquery client: %w", err)
	}

	query := bqClient.Query(fmt.Sprintf(`
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
		`, selectedColumns))

	randomIntSaltPart, ok := ctx.Value(commonhash.RandomIntCtxKey).(int)
	if !ok {
		return nil, fmt.Errorf("unable to get random int salt part from context")
	}

	salt := r.stringSaltPart + strconv.Itoa(randomIntSaltPart)
	query.Parameters = []bigquery.QueryParameter{
		{Name: "emails", Value: args.Emails},
		{Name: "salt", Value: salt},
	}
	rows, err := query.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	var result []GetUserItem
	for {
		var item GetUserItem
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
