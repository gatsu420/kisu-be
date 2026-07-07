package tokenrepo

import (
	"sync"

	"golang.org/x/oauth2"
)

type Repository interface {
	Get(slackUserID string) *oauth2.Token
	Save(slackUserID string, token *oauth2.Token)
}

type repositoryImpl struct {
	mu     sync.RWMutex
	tokens map[string]*oauth2.Token
}

func NewRepository() Repository {
	return &repositoryImpl{
		tokens: map[string]*oauth2.Token{},
	}
}
