package promptanswer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/gatsu420/kisu-be/app/adapter/geminiadapter"
	"github.com/gatsu420/kisu-be/app/middleware"
	"github.com/gatsu420/kisu-be/common/commonhash"
	"golang.org/x/oauth2"
)

type GetAnswerArgs struct {
	Prompt string
	Param  string
}

type GetAnswerResult struct {
	Answer               json.RawMessage `json:"answer"`
	StringifiedFuncCalls string          `json:"stringified_func_calls"`
}

func (u *usecaseImpl) GetAnswer(ctx context.Context, args GetAnswerArgs) (GetAnswerResult, error) {
	hashedParam, err := u.hashParam(ctx, args.Param)
	if err != nil {
		return GetAnswerResult{}, err
	}

	token, ok := ctx.Value(middleware.TokenCtxKey).(*oauth2.Token)
	if !ok {
		return GetAnswerResult{}, fmt.Errorf("token is not found in context")
	}

	content, err := u.geminiAdapter.GetContent(ctx, geminiadapter.GetContentArgs{
		Token:  token,
		Prompt: args.Prompt,
		Param:  hashedParam,
	})
	if err != nil {
		return GetAnswerResult{}, fmt.Errorf("unable to get content from gemini adapter: %w", err)
	}

	return GetAnswerResult{
		Answer:               content.Content,
		StringifiedFuncCalls: content.StringifiedFuncCalls,
	}, nil
}

func (u *usecaseImpl) hashParam(ctx context.Context, param string) (string, error) {
	paramParts := strings.FieldsFunc(param, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	salt, ok := ctx.Value(commonhash.SaltCtxKey).(string)
	if !ok {
		return "", fmt.Errorf("unable to get salt from context")
	}

	hashedParts := commonhash.HashStringSlice(paramParts, salt)
	return strings.Join(hashedParts, ","), nil
}
