package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/app/repository/pgrepo"
	"github.com/gatsu420/kisu-be/common/commonerr"
)

type ctxKey int

const TokenCtxKey ctxKey = iota
const publicErrMsg = "request is unauthorized"

func RefreshToken(pgRepo pgrepo.Repository, googleAuth googleauthadapter.Adapter) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var statusCode int

			cookie, err := r.Cookie("user_id")
			if err != nil {
				statusCode = http.StatusUnauthorized
				slog.Error("unable to find user_id cookie",
					slog.Int(commonerr.StatusCodeKey, statusCode),
					slog.Any(commonerr.ErrKey, err))
				http.Error(w, publicErrMsg, statusCode)
				return
			}

			tokenResult, err := pgRepo.GetUserToken(r.Context(), pgrepo.GetUserTokenArgs{
				UserID: cookie.Value,
			})
			if err != nil {
				statusCode = http.StatusInternalServerError
				slog.Error("unable to get user token",
					slog.Int(commonerr.StatusCodeKey, statusCode),
					slog.Any(commonerr.ErrKey, err))
				http.Error(w, publicErrMsg, statusCode)
				return
			}

			tokenSource := googleAuth.TokenSource(r.Context(), tokenResult.Token)
			freshToken, err := tokenSource.Token()
			if err != nil {
				statusCode = http.StatusInternalServerError
				slog.Error("unable to refresh token",
					slog.Int(commonerr.StatusCodeKey, statusCode),
					slog.Any(commonerr.ErrKey, err))
				http.Error(w, publicErrMsg, statusCode)
				return
			}

			err = pgRepo.AddUserToken(r.Context(), pgrepo.AddUserTokenArgs{
				UserID: cookie.Value,
				Token:  freshToken,
			})
			if err != nil {
				statusCode = http.StatusInternalServerError
				slog.Error("unable to add user token",
					slog.Int(commonerr.StatusCodeKey, statusCode),
					slog.Any(commonerr.ErrKey, err))
				http.Error(w, publicErrMsg, statusCode)
				return
			}

			ctx := context.WithValue(r.Context(), TokenCtxKey, freshToken)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
