package answer

import (
	"context"

	"github.com/gatsu420/kisu-be/app/adapter/driveadapter"
	"github.com/gatsu420/kisu-be/app/adapter/geminiadapter"
)

type Usecase interface {
	RouteTool(ctx context.Context, args RouteToolArgs) (RouteToolResult, error)
	UploadCsv(ctx context.Context, args UploadCsvArgs) (UploadCsvResult, error)
}

type usecaseImpl struct {
	geminiAdapter geminiadapter.Adapter
	driveAdapter  driveadapter.Adapter
}

func NewUsecase(geminiAdapter geminiadapter.Adapter, driveAdapter driveadapter.Adapter) Usecase {
	return &usecaseImpl{
		geminiAdapter: geminiAdapter,
		driveAdapter:  driveAdapter,
	}
}
