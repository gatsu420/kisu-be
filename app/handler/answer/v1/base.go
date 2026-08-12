package answerhandlerv1

import (
	"net/http"

	"github.com/gatsu420/kisu-be/app/usecase/answer"
)

type Handler interface {
	GetAnswer(w http.ResponseWriter, r *http.Request)
}

type handlerImpl struct {
	answerUsecase answer.Usecase
}

func NewHandler(answerUsecase answer.Usecase) Handler {
	return &handlerImpl{
		answerUsecase: answerUsecase,
	}
}
