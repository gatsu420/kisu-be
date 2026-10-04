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
	"golang.org/x/oauth2"
)

type GetAnswerArgs struct {
	Prompt     string
	ParamValue string
	UserID     string
	Limit      string
	Offset     string
}

type GetAnswerResult struct {
	Answer               json.RawMessage
	StringifiedFuncCalls string
}

func (u *usecaseImpl) GetAnswer(ctx context.Context, args GetAnswerArgs) (GetAnswerResult, error) {
	hashedParam, err := u.hashParamValue(ctx, args.ParamValue)
	if err != nil {
		return GetAnswerResult{}, err
	}

	token, ok := ctx.Value(commonctx.TokenCtxKey).(*oauth2.Token)
	if !ok {
		return GetAnswerResult{}, errors.New("unable to get token from context")
	}

	content, err := u.geminiAdapter.GetContent(ctx, geminiadapter.GetContentArgs{
		Token:      token,
		Prompt:     args.Prompt,
		ParamValue: hashedParam,
		UserID:     args.UserID,
		Limit:      args.Limit,
		Offset:     args.Offset,
	})
	if err != nil {
		return GetAnswerResult{}, err
	}

	return GetAnswerResult{
		Answer:               content.Content,
		StringifiedFuncCalls: content.StringifiedTool,
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
