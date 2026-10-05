package answer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/gatsu420/kisu-be/app/adapter/driveadapter"
	"github.com/gatsu420/kisu-be/app/adapter/geminiadapter"
	"github.com/gatsu420/kisu-be/common/commoncrypto"
	"github.com/gatsu420/kisu-be/common/commonctx"
	"github.com/gatsu420/kisu-be/common/commontype"
	"golang.org/x/oauth2"
)

type RouteToolArgs struct {
	Prompt     string
	ParamValue string
	UserID     string
}

type RouteToolResult struct {
	Tool json.RawMessage
	Type commontype.ToolType
}

func (u *usecaseImpl) RouteTool(ctx context.Context, args RouteToolArgs) (RouteToolResult, error) {
	hashedParam, err := u.hashParamValue(ctx, args.ParamValue)
	if err != nil {
		return RouteToolResult{}, err
	}

	content, err := u.geminiAdapter.RouteTool(ctx, geminiadapter.RouteToolArgs{
		Prompt:     args.Prompt,
		ParamValue: hashedParam,
		UserID:     args.UserID,
	})
	if err != nil {
		return RouteToolResult{}, err
	}

	return RouteToolResult{
		Tool: content.Tool,
		Type: content.Type,
	}, nil
}

type UploadCsvArgs struct {
	Name    string
	Content io.Reader
}

type UploadCsvResult struct {
	Url string
}

func (u *usecaseImpl) UploadCsv(ctx context.Context, args UploadCsvArgs) (UploadCsvResult, error) {
	token, ok := ctx.Value(commonctx.TokenCtxKey).(*oauth2.Token)
	if !ok {
		return UploadCsvResult{}, errors.New("unable to get token from context")
	}

	uploadResult, err := u.driveAdapter.UploadCsv(ctx, driveadapter.UploadCsvArgs{
		Name:    args.Name,
		Content: args.Content,
		Token:   token,
	})
	if err != nil {
		return UploadCsvResult{}, err
	}

	return UploadCsvResult{
		Url: uploadResult.Url,
	}, nil
}

func (u *usecaseImpl) hashParamValue(ctx context.Context, value string) (string, error) {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n'
	})
	salt, ok := ctx.Value(commonctx.SaltCtxKey).(string)
	if !ok {
		return "", errors.New("unable to get salt from context")
	}

	hashedParts := commoncrypto.HashStringSlice(parts, salt)
	return strings.Join(hashedParts, ","), nil
}
