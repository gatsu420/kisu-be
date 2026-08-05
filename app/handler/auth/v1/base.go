package authhandlerv1

import (
	"net/http"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/app/repository/staterepo"
	"github.com/gatsu420/kisu-be/app/repository/userrepo"
	"github.com/gatsu420/kisu-be/app/repository/usertokenrepo"
)

type Handler interface {
	GetPermission(w http.ResponseWriter, r *http.Request)
	Callback(w http.ResponseWriter, r *http.Request)
}

type handlerImpl struct {
	googleAuth    googleauthadapter.Adapter
	stateRepo     staterepo.Repository
	userRepo      userrepo.Repository
	userTokenRepo usertokenrepo.Repository
}

func NewHandler(googleAuth googleauthadapter.Adapter, stateRepo staterepo.Repository, userRepo userrepo.Repository, userTokenRepo usertokenrepo.Repository) Handler {
	return &handlerImpl{
		googleAuth:    googleAuth,
		stateRepo:     stateRepo,
		userRepo:      userRepo,
		userTokenRepo: userTokenRepo,
	}
}
