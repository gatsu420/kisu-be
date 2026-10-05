package answerhandlerv1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gatsu420/kisu-be/app/usecase/answer"
	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"github.com/gatsu420/kisu-be/common/commonctx"
	"github.com/gatsu420/kisu-be/common/commonerr"
	"github.com/gatsu420/kisu-be/common/commontype"
	"github.com/google/uuid"
)

type GetAnswerResult struct {
	Answer               json.RawMessage `json:"answer"`
	StringifiedFuncCalls string          `json:"stringified_func_calls"`
}

func (h *handlerImpl) GetAnswer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	userID, ok := r.Context().Value(commonctx.UserIDCtxKey).(*http.Cookie)
	if !ok {
		statusCode = http.StatusUnauthorized
		slog.Error("unable to get user_id cookie",
			slog.Int(commonerr.StatusCodeLogKey, statusCode))
		http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
		return
	}

	// The less uglier way is to construct ctx value as struct, but
	// it's no biggie for now.
	ctx := context.WithValue(r.Context(),
		commonctx.ParamNameCtxKey,
		r.URL.Query().Get("param_name"))
	ctx = context.WithValue(ctx,
		commonctx.SaltCtxKey,
		uuid.New().String())

	promptAnswer, err := h.answerUsecase.GetAnswer(ctx, answer.GetAnswerArgs{
		Prompt:     r.URL.Query().Get("prompt"),
		ParamValue: r.URL.Query().Get("param_value"),
		UserID:     userID.Value,
	})
	if err != nil {
		errMsg = "unable to get answer"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	toolResult, err := h.metadataUsecase.CallTool(ctx, metadata.CallToolArgs{
		Tool:          promptAnswer.Tool,
		Type:          promptAnswer.Type,
		TableLocation: promptAnswer.Tool.Name,
		Limit:         r.URL.Query().Get("limit"),
		Offset:        r.URL.Query().Get("offset"),
	})
	if err != nil {
		errMsg = "unable to call tool"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	err = json.NewEncoder(w).Encode(GetAnswerResult{
		Answer:               toolResult.Result,
		StringifiedFuncCalls: toolResult.StringifiedTool,
	})
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
	}
}

type UploadCsvArgs struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type UploadCsvResult struct {
	Url string `json:"url"`
}

func (h *handlerImpl) UploadCsv(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	var args UploadCsvArgs
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		errMsg = "unable to decode request body"
		statusCode = http.StatusBadRequest
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	result, err := h.answerUsecase.UploadCsv(r.Context(),
		answer.UploadCsvArgs{
			Name:    args.Name,
			Content: strings.NewReader(args.Content),
		})
	if err != nil {
		errMsg = "unable to upload csv"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	err = json.NewEncoder(w).Encode(UploadCsvResult{
		Url: result.Url,
	})
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
	}
}

type AddToolArgs struct {
	ToolDescription string              `json:"tool_description"`
	Project         string              `json:"project"`
	Dataset         string              `json:"dataset"`
	TableName       string              `json:"table_name"`
	Columns         []AddToolColumn     `json:"columns"`
	Type            commontype.ToolType `json:"type"`
	Examples        []AddToolExample    `json:"examples"`
	ParamNames      []string            `json:"param_names"`
}

type AddToolColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type AddToolExample struct {
	Description string `json:"description"`
	Query       string `json:"query"`
}

func (h *handlerImpl) AddTool(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	var args AddToolArgs
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		errMsg = "unable to decode request body"
		statusCode = http.StatusBadRequest
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	columns := []metadata.AddToolColumn{}
	for _, c := range args.Columns {
		columns = append(columns, metadata.AddToolColumn{
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
		})
	}

	examples := []metadata.AddToolExample{}
	for _, e := range args.Examples {
		examples = append(examples, metadata.AddToolExample{
			Description: e.Description,
			Query:       e.Query,
		})
	}

	userID, ok := r.Context().Value(commonctx.UserIDCtxKey).(*http.Cookie)
	if !ok {
		statusCode = http.StatusUnauthorized
		slog.Error("unable to get user_id cookie",
			slog.Int(commonerr.StatusCodeLogKey, statusCode))
		http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
		return
	}

	err = h.metadataUsecase.AddTool(r.Context(), metadata.AddToolArgs{
		UserID:          userID.Value,
		ToolDescription: args.ToolDescription,
		Project:         args.Project,
		Dataset:         args.Dataset,
		TableName:       args.TableName,
		Columns:         columns,
		Type:            args.Type,
		Examples:        examples,
		ParamNames:      args.ParamNames,
	})
	if err != nil {
		errMsg = "unable to add tool"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("tool is added"))
}

type GetToolResult struct {
	Rows []GetToolRow `json:"rows"`
}

type GetToolRow struct {
	ToolDescription string              `json:"tool_description"`
	Project         string              `json:"project"`
	Dataset         string              `json:"dataset"`
	TableName       string              `json:"table_name"`
	Columns         []GetToolColumn     `json:"columns"`
	Type            commontype.ToolType `json:"type"`
	Examples        []GetToolExample    `json:"examples"`
	ParamNames      []string            `json:"param_names"`
}

type GetToolColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type GetToolExample struct {
	Description string `json:"description"`
	Query       string `json:"query"`
}

func (h *handlerImpl) GetTool(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	userID, ok := r.Context().Value(commonctx.UserIDCtxKey).(*http.Cookie)
	if !ok {
		statusCode = http.StatusUnauthorized
		slog.Error("unable to get user_id cookie",
			slog.Int(commonerr.StatusCodeLogKey, statusCode))
		http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
		return
	}

	tools, err := h.metadataUsecase.GetTool(r.Context(), metadata.GetToolArgs{
		UserID: userID.Value,
	})
	if err != nil {
		errMsg = "unable to get tool"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	resultRows := []GetToolRow{}
	for _, r := range tools.Rows {
		columns := []GetToolColumn{}
		for _, c := range r.Columns {
			columns = append(columns, GetToolColumn{
				Name:        c.Name,
				Type:        c.Type,
				Description: c.Description,
			})
		}

		examples := []GetToolExample{}
		for _, e := range r.Examples {
			examples = append(examples, GetToolExample{
				Description: e.Description,
				Query:       e.Query,
			})
		}

		resultRows = append(resultRows, GetToolRow{
			ToolDescription: r.ToolDescription,
			Project:         r.Project,
			Dataset:         r.Dataset,
			TableName:       r.TableName,
			Columns:         columns,
			Type:            r.Type,
			Examples:        examples,
			ParamNames:      r.ParamNames,
		})
	}

	err = json.NewEncoder(w).Encode(resultRows)
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
	}
}

type ValidateToolQueryResult struct {
	IsValid bool `json:"is_valid"`
}

func (h *handlerImpl) ValidateToolQuery(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	result, err := h.metadataUsecase.ValidateToolQuery(r.Context(), metadata.ValidateToolQueryArgs{
		Query: r.URL.Query().Get("query"),
	})
	if err != nil {
		errMsg = "unable to validate tool query"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	err = json.NewEncoder(w).Encode(ValidateToolQueryResult{
		IsValid: result.IsValid,
	})
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
	}
}

type GetToolTableMetadataResult struct {
	Description string                       `json:"description"`
	Columns     []GetToolTableMetadataColumn `json:"columns"`
}

type GetToolTableMetadataColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func (h *handlerImpl) GetToolTableMetadata(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	result, err := h.metadataUsecase.GetToolTableMetadata(r.Context(), metadata.GetToolTableMetadataArgs{
		Type:         commontype.ToolType(r.URL.Query().Get("type")),
		Project:      r.URL.Query().Get("project"),
		Dataset:      r.URL.Query().Get("dataset"),
		TableName:    r.URL.Query().Get("table_name"),
		BuilderQuery: r.URL.Query().Get("builder_query"),
	})
	if err != nil {
		errMsg = "unable to get tool table metadata"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	columns := []GetToolTableMetadataColumn{}
	for _, c := range result.Columns {
		columns = append(columns, GetToolTableMetadataColumn{
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
		})
	}

	err = json.NewEncoder(w).Encode(GetToolTableMetadataResult{
		Description: result.Description,
		Columns:     columns,
	})
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
	}
}
