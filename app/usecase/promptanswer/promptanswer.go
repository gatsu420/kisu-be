package promptanswer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gatsu420/kisu-be/app/adapter/geminiadapter"
	"github.com/gatsu420/kisu-be/common/commonhash"
)

type GetAnswerArgs struct {
	HashedEmail string
	Prompt      string
	Param       string
}

type GetAnswerResult struct {
	Answer json.RawMessage `json:"answer"`
}

func (u *usecaseImpl) GetAnswer(ctx context.Context, args GetAnswerArgs) (GetAnswerResult, error) {
	hashedParam, err := u.hashParam(ctx, args.Param)
	if err != nil {
		return GetAnswerResult{}, err
	}

	token := u.tokenRepo.Get(args.HashedEmail)
	content, err := u.geminiAdapter.GetContent(ctx, geminiadapter.GetContentArgs{
		Token:  token,
		Prompt: args.Prompt,
		Param:  hashedParam,
	})
	if err != nil {
		return GetAnswerResult{}, fmt.Errorf("unable to get content from gemini adapter: %w", err)
	}

	return GetAnswerResult{
		Answer: content.Content,
	}, nil
}

func (u *usecaseImpl) hashParam(ctx context.Context, param string) (string, error) {
	paramParts := strings.Split(param, "\n")
	randomIntSaltPart, ok := ctx.Value(commonhash.RandomIntCtxKey).(int)
	if !ok {
		return "", fmt.Errorf("unable to get random int salt part from context")
	}

	salt := u.stringSaltPart + strconv.Itoa(randomIntSaltPart)
	hashedParts := commonhash.HashStringSlice(paramParts, salt)
	return strings.Join(hashedParts, ","), nil
}
