package r2repo

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type PutObjectArgs struct {
	AccessKeyID     string
	AccessKeySecret string
	AccountID       string
	Bucket          string
	Key             string
	Body            []byte
}

func (r *repositoryImpl) PutObject(ctx context.Context, args PutObjectArgs) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(args.AccessKeyID, args.AccessKeySecret, "")),
		config.WithRegion("auto"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%v.r2.cloudflarestorage.com", args.AccountID))
	})

	input := &s3.PutObjectInput{
		Bucket:      aws.String(args.Bucket),
		Key:         aws.String(args.Key),
		Body:        bytes.NewReader(args.Body),
		ContentType: aws.String("text/yaml"),
	}

	_, err = client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	return nil
}
