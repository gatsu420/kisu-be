package authhandlerv1

import (
	"net/http"

	"github.com/gatsu420/kisu/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu/app/usecase/metadata"
)

type Handler interface {
	GetPermission(w http.ResponseWriter, r *http.Request)
	Callback(w http.ResponseWriter, r *http.Request)
}

type handlerImpl struct {
	hashSecret      string
	googleAuth      googleauthadapter.Adapter
	metadataUsecase metadata.Usecase
}

func NewHandler(hashSecret string, googleAuth googleauthadapter.Adapter, metadataUsecase metadata.Usecase) Handler {
	return &handlerImpl{
		hashSecret:      hashSecret,
		googleAuth:      googleAuth,
		metadataUsecase: metadataUsecase,
	}
}
