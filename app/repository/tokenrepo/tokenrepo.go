package tokenrepo

import (
	"fmt"

	"golang.org/x/oauth2"
)

var ErrTokenNotFound = fmt.Errorf("token is not found")

func (r *repositoryImpl) Get(email string) *oauth2.Token {
	r.mu.Lock()
	defer r.mu.Unlock()

	token, ok := r.tokens[email]
	if !ok {
		return nil
	}

	return token
}

func (r *repositoryImpl) Save(email string, token *oauth2.Token) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tokens[email] = token
}
