package answerhandlerv1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

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
		commonctx.FilterCtxKey,
		r.URL.Query().Get("filter"))
	ctx = context.WithValue(ctx,
		commonctx.SaltCtxKey,
		uuid.New().String())

	prompt := r.URL.Query().Get("prompt")
	param := r.URL.Query().Get("param")
	promptAnswer, err := h.answerUsecase.GetAnswer(ctx, answer.GetAnswerArgs{
		Prompt: prompt,
		Param:  param,
		UserID: userID.Value,
	})
	if err != nil {
		errMsg = "unable to get answer"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.Any(commonerr.ErrLogKey, err))
		http.Error(w, errMsg, statusCode)
		return
	}

	err = json.NewEncoder(w).Encode(GetAnswerResult{
		Answer:               promptAnswer.Answer,
		StringifiedFuncCalls: promptAnswer.StringifiedFuncCalls,
	})
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.Any(commonerr.ErrLogKey, err))
		http.Error(w, errMsg, statusCode)
	}
}

type AddToolArgs struct {
	ToolDescription  string              `json:"tool_description"`
	Project          string              `json:"project"`
	Dataset          string              `json:"dataset"`
	TableName        string              `json:"table_name"`
	Columns          []AddToolColumn     `json:"columns"`
	Type             commontype.ToolType `json:"type"`
	Examples         []AddToolExample    `json:"examples"`
	ParamName        string              `json:"param_name"`
	ParamType        string              `json:"param_type"`
	ParamDescription string              `json:"param_description"`
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
			slog.Any(commonerr.ErrLogKey, err))
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
		UserID:           userID.Value,
		ToolDescription:  args.ToolDescription,
		Project:          args.Project,
		Dataset:          args.Dataset,
		TableName:        args.TableName,
		Columns:          columns,
		Type:             args.Type,
		Examples:         examples,
		ParamName:        args.ParamName,
		ParamType:        args.ParamType,
		ParamDescription: args.ParamDescription,
	})
	if err != nil {
		errMsg = "unable to add tool"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.Any(commonerr.ErrLogKey, err))
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
	ToolDescription  string              `json:"tool_description"`
	Project          string              `json:"project"`
	Dataset          string              `json:"dataset"`
	TableName        string              `json:"table_name"`
	Columns          []GetToolColumn     `json:"columns"`
	Type             commontype.ToolType `json:"type"`
	Examples         []GetToolExample    `json:"examples"`
	ParamName        string              `json:"param_name"`
	ParamType        string              `json:"param_type"`
	ParamDescription string              `json:"param_description"`
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
			slog.Any(commonerr.ErrLogKey, err))
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
			ToolDescription:  r.ToolDescription,
			Project:          r.Project,
			Dataset:          r.Dataset,
			TableName:        r.TableName,
			Columns:          columns,
			Type:             r.Type,
			Examples:         examples,
			ParamName:        r.ParamName,
			ParamType:        r.ParamType,
			ParamDescription: r.ParamDescription,
		})
	}

	err = json.NewEncoder(w).Encode(resultRows)
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.Any(commonerr.ErrLogKey, err))
		http.Error(w, errMsg, statusCode)
	}
}
