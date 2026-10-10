package answer

import (
	"context"

	"github.com/gatsu420/kisu/app/adapter/driveadapter"
	"github.com/gatsu420/kisu/app/adapter/geminiadapter"
	"github.com/gatsu420/kisu/app/repository/pgrepo"
)

type Usecase interface {
	RouteTool(ctx context.Context, args RouteToolArgs) (RouteToolResult, error)
	UploadCsv(ctx context.Context, args UploadCsvArgs) (UploadCsvResult, error)
	AddBookmark(ctx context.Context, args AddBookmarkArgs) error
	ListBookmark(ctx context.Context, args ListBookmarkArgs) (ListBookmarkResult, error)
	GetBookmark(ctx context.Context, args GetBookmarkArgs) (GetBookmarkResult, error)
	DeleteBookmark(ctx context.Context, args DeleteBookmarkArgs) error
}

type usecaseImpl struct {
	geminiAdapter geminiadapter.Adapter
	driveAdapter  driveadapter.Adapter
	pgRepo        pgrepo.Repository
}

func NewUsecase(geminiAdapter geminiadapter.Adapter, driveAdapter driveadapter.Adapter, pgRepo pgrepo.Repository) Usecase {
	return &usecaseImpl{
		geminiAdapter: geminiAdapter,
		driveAdapter:  driveAdapter,
		pgRepo:        pgRepo,
	}
}
