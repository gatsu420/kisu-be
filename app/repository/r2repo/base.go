package r2repo

import "context"

type Repository interface {
	PutObject(ctx context.Context, args PutObjectArgs) error
}

type repositoryImpl struct{}

func NewRepository() Repository {
	return &repositoryImpl{}
}
