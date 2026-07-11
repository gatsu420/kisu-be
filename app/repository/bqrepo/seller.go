package bqrepo

import (
	"context"
	"fmt"
	"strconv"

	"cloud.google.com/go/bigquery"
	"github.com/gatsu420/kisu-be/common/commonhash"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GetSellerArgs struct {
	Token *oauth2.Token
	Query string
}

func (r *repositoryImpl) GetSeller(ctx context.Context, args GetSellerArgs) ([]map[string]bigquery.Value, error) {
	googleAuthClient := r.googleAuth.Client(ctx, args.Token)
	bqClient, err := bigquery.NewClient(ctx, r.projectID, option.WithHTTPClient(googleAuthClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create bigquery client: %w", err)
	}

	randomIntSaltPart, ok := ctx.Value(commonhash.RandomIntCtxKey).(int)
	if !ok {
		return nil, fmt.Errorf("unable to get random int salt part from context")
	}

	salt := r.stringSaltPart + strconv.Itoa(randomIntSaltPart)
	hashQuery := bqClient.Query(fmt.Sprintf(`
		create or replace view rumah-aya.some_event.merchants_hash_query as
		select
			* except(email),
			to_base64(sha256(concat(email, "%v"))) hashed_email
		from rumah-aya.some_event.merchants
		`, salt))

	job, err := hashQuery.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to run job for hash query: %w", err)
	}

	jobStatus, err := job.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to finish job for hash query: %w", err)
	}
	if jobStatus.Err() != nil {
		return nil, fmt.Errorf("hash query job finished with error: %w", jobStatus.Err())
	}

	getterQuery := bqClient.Query(args.Query)
	rows, err := getterQuery.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to run job for getter query: %w", err)
	}

	result := []map[string]bigquery.Value{}
	for {
		var item map[string]bigquery.Value
		err := rows.Next(&item)
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("row doesn't conform to item struct: %w", err)
		}

		result = append(result, item)
	}

	return result, nil
}
