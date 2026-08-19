package answerhandlerv1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gatsu420/kisu-be/app/usecase/answer"
	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"github.com/gatsu420/kisu-be/common/commonerr"
	"github.com/gatsu420/kisu-be/common/commonhash"
	"github.com/google/uuid"
)

type GetAnswerArgs struct {
	Prompt string `json:"prompt"`
	Param  string `json:"param"`
}

func (h *handlerImpl) GetAnswer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	salt := uuid.New().String()
	ctx := context.WithValue(r.Context(), commonhash.SaltCtxKey, salt)
	prompt := r.URL.Query().Get("prompt")
	param := r.URL.Query().Get("param")
	promptAnswer, err := h.answerUsecase.GetAnswer(ctx, answer.GetAnswerArgs{
		Prompt: prompt,
		Param:  param,
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
	ToolDescription string                `json:"tool_description"`
	TableName       string                `json:"table_name"`
	Columns         []AddToolColumn       `json:"columns"`
	QueryExamples   []AddToolQueryExample `json:"query_examples"`
}

type AddToolColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	IsSelected  bool   `json:"is_selected"`
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
			IsSelected:  c.IsSelected,
		})
	}

	queryExamples := []metadata.AddToolQueryExample{}
	for _, qe := range args.QueryExamples {
		queryExamples = append(queryExamples, metadata.AddToolQueryExample{
			Description: qe.Description,
			Query:       qe.Query,
		})
	}

	userID, err := r.Cookie("user_id")
	if err != nil {
		errMsg = "unable to get user_id cookie"
		statusCode = http.StatusUnauthorized
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeKey, statusCode),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, statusCode)
		return
	}

	err = h.metadataUsecase.AddTool(r.Context(), metadata.AddToolArgs{
		UserID:          userID.Value,
		ToolDescription: args.ToolDescription,
		TableName:       args.TableName,
		Columns:         columns,
		QueryExamples:   queryExamples,
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
