package answer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/gatsu420/kisu/app/adapter/driveadapter"
	"github.com/gatsu420/kisu/app/adapter/geminiadapter"
	"github.com/gatsu420/kisu/app/repository/pgrepo"
	"github.com/gatsu420/kisu/common/commoncrypto"
	"github.com/gatsu420/kisu/common/commonctx"
	"github.com/gatsu420/kisu/common/commontype"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type RouteToolArgs struct {
	Prompt     string
	ParamName  string
	ParamValue string
	Salt       string
	UserID     string
}

type RouteToolResult struct {
	Tool json.RawMessage
	Type commontype.ToolType
}

func (u *usecaseImpl) RouteTool(ctx context.Context, args RouteToolArgs) (RouteToolResult, error) {
	hashedParam, err := u.hashParamValue(hashParamValueArgs{
		value: args.ParamValue,
		salt:  args.Salt,
	})
	if err != nil {
		return RouteToolResult{}, err
	}

	content, err := u.geminiAdapter.RouteTool(ctx, geminiadapter.RouteToolArgs{
		Prompt:     args.Prompt,
		ParamName:  args.ParamName,
		ParamValue: hashedParam.value,
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

type AddBookmarkArgs struct {
	ID         string
	UserID     string
	Name       string
	ParamName  string
	ParamValue string
	Query      string
	HashedTool string
}

func (u *usecaseImpl) AddBookmark(ctx context.Context, args AddBookmarkArgs) error {
	var id string
	if args.ID == "" {
		id = uuid.New().String()
	} else {
		id = args.ID
	}

	nameRunes := []rune(args.Name)
	if len(nameRunes) > 30 {
		args.Name = string(nameRunes[:30]) + "..."
	}

	return u.pgRepo.AddBookmark(ctx, pgrepo.AddBookmarkArgs{
		ID:         id,
		UserID:     args.UserID,
		Name:       args.Name,
		ParamName:  args.ParamName,
		ParamValue: args.ParamValue,
		Query:      args.Query,
		HashedTool: args.HashedTool,
	})
}

type ListBookmarkArgs struct {
	UserID string
}

type ListBookmarkResult struct {
	Rows []ListBookmarkRow
}

type ListBookmarkRow struct {
	ID        string
	Name      string
	UpdatedAt time.Time
}

func (u *usecaseImpl) ListBookmark(ctx context.Context, args ListBookmarkArgs) (ListBookmarkResult, error) {
	rows, err := u.pgRepo.ListBookmark(ctx, pgrepo.ListBookmarkArgs{
		UserID: args.UserID,
	})
	if err != nil {
		return ListBookmarkResult{}, err
	}

	resultRows := make([]ListBookmarkRow, len(rows.Rows))
	for i, v := range rows.Rows {
		resultRows[i] = ListBookmarkRow{
			ID:        v.ID,
			Name:      v.Name,
			UpdatedAt: v.UpdatedAt,
		}
	}

	return ListBookmarkResult{
		Rows: resultRows,
	}, nil
}

type GetBookmarkArgs struct {
	ID     string
	UserID string
}

type GetBookmarkResult struct {
	Name       string
	ParamName  string
	ParamValue string
	Query      string
	HashedTool string
	UpdatedAt  time.Time
}

func (u *usecaseImpl) GetBookmark(ctx context.Context, args GetBookmarkArgs) (GetBookmarkResult, error) {
	rows, err := u.pgRepo.GetBookmark(ctx, pgrepo.GetBookmarkArgs{
		ID:     args.ID,
		UserID: args.UserID,
	})
	if err != nil {
		return GetBookmarkResult{}, err
	}

	return GetBookmarkResult{
		Name:       rows.Name,
		ParamName:  rows.ParamName,
		ParamValue: rows.ParamValue,
		Query:      rows.Query,
		HashedTool: rows.HashedTool,
		UpdatedAt:  rows.UpdatedAt,
	}, nil
}

type DeleteBookmarkArgs struct {
	ID string
}

func (u *usecaseImpl) DeleteBookmark(ctx context.Context, args DeleteBookmarkArgs) error {
	return u.pgRepo.DeleteBookmark(ctx, pgrepo.DeleteBookmarkArgs{
		ID: args.ID,
	})
}

type hashParamValueArgs struct {
	value string
	salt  string
}

type hashParamValueResult struct {
	value string
}

func (u *usecaseImpl) hashParamValue(args hashParamValueArgs) (hashParamValueResult, error) {
	parts := strings.FieldsFunc(args.value, func(r rune) bool {
		return r == ',' || r == '\n'
	})
	hashedParts := commoncrypto.HashStringSlice(parts, args.salt)

	return hashParamValueResult{
		value: strings.Join(hashedParts, ","),
	}, nil
}
