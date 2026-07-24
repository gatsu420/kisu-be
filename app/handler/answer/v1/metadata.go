package answerhandlerv1

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gatsu420/kisu-be/app/usecase/promptanswer"
	"github.com/gatsu420/kisu-be/common/commonerr"
)

type Column struct {
	Name        string
	Type        string
	Description string
}

type QueryExample struct {
	Description string
	Query       string
}

type AddMetadataReqBody struct {
	ToolDescription string
	TableName       string
	Columns         []Column
	QueryExamples   []QueryExample
}

func (h *handlerImpl) AddMetadata(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	_, err := r.Cookie("hashed_email")
	if err != nil {
		errMsg = "login session is expired"
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeKey, http.StatusUnauthorized),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	var reqBody AddMetadataReqBody
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		errMsg = "unable to decode request body"
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeKey, http.StatusBadRequest),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	columns := []promptanswer.Column{}
	for _, c := range reqBody.Columns {
		columns = append(columns, promptanswer.Column{
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
		})
	}

	queryExamples := []promptanswer.QueryExample{}
	for _, e := range reqBody.QueryExamples {
		queryExamples = append(queryExamples, promptanswer.QueryExample{
			Description: e.Description,
			Query:       e.Query,
		})
	}

	err = h.promptAnswerUsecase.AddMetadata(r.Context(), promptanswer.AddMetadataArgs{
		ToolDescription: reqBody.ToolDescription,
		TableName:       reqBody.TableName,
		Columns:         columns,
		QueryExamples:   queryExamples,
	})
	if err != nil {
		errMsg = "unable to add metadata"
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeKey, http.StatusBadRequest),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("metadata is added"))
}
