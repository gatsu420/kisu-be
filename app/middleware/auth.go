package middleware

import (
	"context"
	"net/http"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/app/repository/pgrepo"
)

type ctxKey int

const TokenCtxKey ctxKey = iota

func RefreshToken(pgRepo pgrepo.Repository, googleAuth googleauthadapter.Adapter) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("user_id")
			if err != nil {
				http.Error(w, "request is unauthorized", http.StatusUnauthorized)
				return
			}

			tokenResult, err := pgRepo.GetUserToken(r.Context(), pgrepo.GetUserTokenArgs{
				UserID: cookie.Value,
			})
			if err != nil {
				http.Error(w, "request is unauthorized", http.StatusUnauthorized)
				return
			}

			tokenSource := googleAuth.TokenSource(r.Context(), tokenResult.Token)
			freshToken, err := tokenSource.Token()
			if err != nil {
				http.Error(w, "unable to refresh token", http.StatusInternalServerError)
				return
			}

			err = pgRepo.InsertUserToken(r.Context(), pgrepo.InsertUserTokenArgs{
				UserID: cookie.Value,
				Token:  freshToken,
			})
			if err != nil {
				http.Error(w, "unable to insert user token", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), TokenCtxKey, freshToken)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
