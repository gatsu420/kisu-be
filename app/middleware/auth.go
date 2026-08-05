package middleware

import (
	"context"
	"net/http"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/app/repository/usertokenrepo"
)

type contextKey string

const TokenCtxKey contextKey = "token"

func TokenRefresh(userTokenRepo usertokenrepo.Repository, googleAuth googleauthadapter.Adapter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Get user_id from cookie
			userIDCookie, err := r.Cookie("user_id")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// 2. Get token from DB
			token, err := userTokenRepo.Get(r.Context(), userIDCookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// 3. Create token source (handles refresh automatically)
			tokenSource := googleAuth.TokenSource(r.Context(), token)

			// 4. Get fresh token (refreshes if expired)
			freshToken, err := tokenSource.Token()
			if err != nil {
				http.Error(w, "token refresh failed", http.StatusInternalServerError)
				return
			}

			// 5. Save refreshed token back to DB
			err = userTokenRepo.Save(r.Context(), userIDCookie.Value, freshToken)
			if err != nil {
				http.Error(w, "failed to save token", http.StatusInternalServerError)
				return
			}

			// 6. Put fresh token in context
			ctx := context.WithValue(r.Context(), TokenCtxKey, freshToken)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
