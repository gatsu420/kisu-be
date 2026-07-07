package authhandlerv1

import (
	"net/http"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/app/repository/staterepo"
	"github.com/gatsu420/kisu-be/app/repository/tokenrepo"
)

type Handler interface {
	GetPermission(w http.ResponseWriter, r *http.Request)
	Callback(w http.ResponseWriter, r *http.Request)
}

type handlerImpl struct {
	googleAuth     googleauthadapter.Adapter
	stateRepo      staterepo.Repository
	tokenRepo      tokenrepo.Repository
	secret         string
	stringSaltPart string
}

func NewHandler(googleAuth googleauthadapter.Adapter, stateRepo staterepo.Repository, tokenRepo tokenrepo.Repository, secret string, stringSaltPart string) Handler {
	return &handlerImpl{
		googleAuth:     googleAuth,
		stateRepo:      stateRepo,
		tokenRepo:      tokenRepo,
		secret:         secret,
		stringSaltPart: stringSaltPart,
	}
}
