package answerhandlerv1

import (
	"net/http"

	"github.com/gatsu420/kisu-be/app/usecase/answer"
	"github.com/gatsu420/kisu-be/app/usecase/metadata"
)

type Handler interface {
	AddTool(w http.ResponseWriter, r *http.Request)
	GetAnswer(w http.ResponseWriter, r *http.Request)
}

type handlerImpl struct {
	metadataUsecase metadata.Usecase
	answerUsecase   answer.Usecase
}

func NewHandler(metadataUsecase metadata.Usecase, answerUsecase answer.Usecase) Handler {
	return &handlerImpl{
		metadataUsecase: metadataUsecase,
		answerUsecase:   answerUsecase,
	}
}
