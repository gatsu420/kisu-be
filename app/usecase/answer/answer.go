package answer

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gatsu420/kisu-be/app/adapter/geminiadapter"
	"github.com/gatsu420/kisu-be/common/commoncrypto"
	"github.com/gatsu420/kisu-be/common/commonctx"
	"golang.org/x/oauth2"
)

type GetAnswerArgs struct {
	Prompt string
	Param  string
	UserID string
}

type GetAnswerResult struct {
	Answer               json.RawMessage
	StringifiedFuncCalls string
}

func (u *usecaseImpl) GetAnswer(ctx context.Context, args GetAnswerArgs) (GetAnswerResult, error) {
	hashedParam, err := u.hashParam(ctx, args.Param)
	if err != nil {
		return GetAnswerResult{}, err
	}

	token, ok := ctx.Value(commonctx.TokenCtxKey).(*oauth2.Token)
	if !ok {
		return GetAnswerResult{}, errors.New("unable to get token from context")
	}

	content, err := u.geminiAdapter.GetContent(ctx, geminiadapter.GetContentArgs{
		Token:  token,
		Prompt: args.Prompt,
		Param:  hashedParam,
		UserID: args.UserID,
	})
	if err != nil {
		return GetAnswerResult{}, err
	}

	return GetAnswerResult{
		Answer:               content.Content,
		StringifiedFuncCalls: content.StringifiedTool,
	}, nil
}

func (u *usecaseImpl) hashParam(ctx context.Context, param string) (string, error) {
	paramParts := strings.FieldsFunc(param, func(r rune) bool {
		return r == ',' || r == '\n'
	})
	salt, ok := ctx.Value(commonctx.SaltCtxKey).(string)
	if !ok {
		return "", errors.New("unable to get salt from context")
	}

	hashedParts := commoncrypto.HashStringSlice(paramParts, salt)
	return strings.Join(hashedParts, ","), nil
}
