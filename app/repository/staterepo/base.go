package staterepo

import "sync"

type Repository interface {
	Save(state string)
	CheckExistence(state string) bool
}

type repositoryImpl struct {
	mu     sync.RWMutex
	states map[string]struct{}
}

func NewRepository() Repository {
	return &repositoryImpl{
		states: map[string]struct{}{},
	}
}
