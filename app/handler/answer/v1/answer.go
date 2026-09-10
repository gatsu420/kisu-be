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
	"github.com/google/uuid"
)

func (h *handlerImpl) GetAnswer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int
	userID, ok := r.Context().Value(commonctx.UserIDCtxKey).(*http.Cookie)
	if !ok {
		statusCode = http.StatusUnauthorized
		slog.Error("unable to get user_id cookie",
			slog.Int(commonerr.StatusCodeKey, statusCode))
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
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(promptAnswer)
	if err != nil {
		errMsg = "unable to write response"
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, http.StatusBadRequest)
	}
}

type AddToolArgs struct {
	ToolDescription  string                `json:"tool_description"`
	TableName        string                `json:"table_name"`
	Columns          []AddToolColumn       `json:"columns"`
	QueryExamples    []AddToolQueryExample `json:"query_examples"`
	ParamName        string                `json:"param_name"`
	ParamType        string                `json:"param_type"`
	ParamDescription string                `json:"param_description"`
}

type AddToolColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type AddToolQueryExample struct {
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
			slog.Int(commonerr.StatusCodeKey, statusCode),
			slog.Any(commonerr.ErrKey, err))
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

	queryExamples := []metadata.AddToolQueryExample{}
	for _, qe := range args.QueryExamples {
		queryExamples = append(queryExamples, metadata.AddToolQueryExample{
			Description: qe.Description,
			Query:       qe.Query,
		})
	}

	userID, ok := r.Context().Value(commonctx.UserIDCtxKey).(*http.Cookie)
	if !ok {
		statusCode = http.StatusUnauthorized
		slog.Error("unable to get user_id cookie",
			slog.Int(commonerr.StatusCodeKey, statusCode))
		http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
		return
	}

	err = h.metadataUsecase.AddTool(r.Context(), metadata.AddToolArgs{
		UserID:           userID.Value,
		ToolDescription:  args.ToolDescription,
		TableName:        args.TableName,
		Columns:          columns,
		QueryExamples:    queryExamples,
		ParamName:        args.ParamName,
		ParamType:        args.ParamType,
		ParamDescription: args.ParamDescription,
	})
	if err != nil {
		errMsg = "unable to add tool"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeKey, statusCode),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("tool is added"))
}
