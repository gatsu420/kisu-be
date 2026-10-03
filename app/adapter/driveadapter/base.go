package driveadapter

import (
	"context"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
)

type Adapter interface {
	UploadCsv(ctx context.Context, args UploadCsvArgs) (UploadCsvResult, error)
}

type adapterImpl struct {
	googleAuthAdapter googleauthadapter.Adapter
}

func NewAdapter(googleAuthAdapter googleauthadapter.Adapter) Adapter {
	return &adapterImpl{
		googleAuthAdapter: googleAuthAdapter,
	}
}
