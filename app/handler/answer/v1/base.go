package answerhandlerv1

import (
	"net/http"

	"github.com/gatsu420/kisu/app/usecase/answer"
	"github.com/gatsu420/kisu/app/usecase/metadata"
)

type Handler interface {
	AddTool(w http.ResponseWriter, r *http.Request)
	GetTool(w http.ResponseWriter, r *http.Request)
	ValidateToolQuery(w http.ResponseWriter, r *http.Request)
	GetToolTableMetadata(w http.ResponseWriter, r *http.Request)
	RouteTool(w http.ResponseWriter, r *http.Request)
	GetAnswer(w http.ResponseWriter, r *http.Request)
	UploadCsv(w http.ResponseWriter, r *http.Request)
	AddBookmark(w http.ResponseWriter, r *http.Request)
	ListBookmark(w http.ResponseWriter, r *http.Request)
	GetBookmark(w http.ResponseWriter, r *http.Request)
	DeleteBookmark(w http.ResponseWriter, r *http.Request)
}

type handlerImpl struct {
	hashSecret      string
	metadataUsecase metadata.Usecase
	answerUsecase   answer.Usecase
}

func NewHandler(hashSecret string, metadataUsecase metadata.Usecase, answerUsecase answer.Usecase) Handler {
	return &handlerImpl{
		hashSecret:      hashSecret,
		metadataUsecase: metadataUsecase,
		answerUsecase:   answerUsecase,
	}
}
