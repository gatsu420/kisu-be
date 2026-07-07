package answerhandlerv1

import (
	"net/http"

	"github.com/gatsu420/kisu-be/app/usecase/promptanswer"
)

type Handler interface {
	GetAnswer(w http.ResponseWriter, r *http.Request)
}

type handlerImpl struct {
	promptAnswerUsecase promptanswer.Usecase
}

func NewHandler(promptAnswerUsecase promptanswer.Usecase) Handler {
	return &handlerImpl{
		promptAnswerUsecase: promptAnswerUsecase,
	}
}
