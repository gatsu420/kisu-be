package authhandlerv1

import (
	"net/http"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/app/usecase/metadata"
)

type Handler interface {
	GetPermission(w http.ResponseWriter, r *http.Request)
	Callback(w http.ResponseWriter, r *http.Request)
}

type handlerImpl struct {
	googleAuth      googleauthadapter.Adapter
	metadataUsecase metadata.Usecase
}

func NewHandler(googleAuth googleauthadapter.Adapter, metadataUsecase metadata.Usecase) Handler {
	return &handlerImpl{
		googleAuth:      googleAuth,
		metadataUsecase: metadataUsecase,
	}
}
