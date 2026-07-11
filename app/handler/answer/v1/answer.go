package answerhandlerv1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gatsu420/kisu-be/app/usecase/promptanswer"
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

	var args GetAnswerArgs
	var errMsg string
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		errMsg = "unable to decode request body"
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeKey, http.StatusBadRequest),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	hashedEmail, err := r.Cookie("hashed_email")
	if err != nil {
		errMsg = "login session is expired"
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeKey, http.StatusUnauthorized),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	salt := uuid.New().String()
	ctx := context.WithValue(r.Context(), commonhash.SaltCtxKey, salt)
	answer, err := h.promptAnswerUsecase.GetAnswer(ctx, promptanswer.GetAnswerArgs{
		HashedEmail: hashedEmail.Value,
		Prompt:      args.Prompt,
		Param:       args.Param,
	})
	if err != nil {
		errMsg = "unable to get answer"
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(answer)
	if err != nil {
		errMsg = "unable to write response"
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, http.StatusBadRequest)
	}
}
