package driveadapter

import (
	"context"
	"fmt"
	"io"

	"github.com/gatsu420/kisu/app/adapter/googleauthadapter"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type UploadCsvArgs struct {
	Name    string
	Content io.Reader
	Token   *oauth2.Token
}

type UploadCsvResult struct {
	Url string
}

func (a *adapterImpl) UploadCsv(ctx context.Context, args UploadCsvArgs) (UploadCsvResult, error) {
	googleAuthClient := a.googleAuthAdapter.Client(ctx,
		googleauthadapter.ClientArgs{
			Token: args.Token,
		})
	driveService, err := drive.NewService(ctx,
		option.WithHTTPClient(googleAuthClient.Client))
	if err != nil {
		return UploadCsvResult{}, fmt.Errorf("unable to create drive service: %w", err)
	}

	mimeType := "text/csv"
	file, err := driveService.Files.Create(&drive.File{
		Name:     args.Name,
		MimeType: mimeType,
	}).Media(args.Content, googleapi.ContentType(mimeType)).
		Fields(googleapi.Field("webViewLink")).
		Do()
	if err != nil {
		return UploadCsvResult{}, fmt.Errorf("unable to create file in drive: %w", err)
	}

	return UploadCsvResult{
		Url: file.WebViewLink,
	}, nil
}
