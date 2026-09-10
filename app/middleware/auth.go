package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"github.com/gatsu420/kisu-be/common/commonerr"
)

type ctxKey int

const (
	UserIDCtxKey ctxKey = iota
	TokenCtxKey
)

func RefreshToken(metadataUsecase metadata.Usecase, googleAuth googleauthadapter.Adapter) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var statusCode int

			userID, err := r.Cookie("user_id")
			if err != nil {
				statusCode = http.StatusUnauthorized
				slog.Error("unable to find user_id cookie",
					slog.Int(commonerr.StatusCodeKey, statusCode),
					slog.Any(commonerr.ErrKey, err))
				http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDCtxKey, userID)

			token, err := metadataUsecase.GetUserToken(ctx, metadata.GetUserTokenArgs{
				UserID: userID.Value,
			})
			if err != nil {
				statusCode = http.StatusInternalServerError
				slog.Error("unable to get user token",
					slog.Int(commonerr.StatusCodeKey, statusCode),
					slog.Any(commonerr.ErrKey, err))
				http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
				return
			}

			tokenSource := googleAuth.TokenSource(ctx, googleauthadapter.TokenSourceArgs{
				Token: token.Token,
			})
			freshToken, err := tokenSource.Source.Token()
			if err != nil {
				statusCode = http.StatusInternalServerError
				slog.Error("unable to refresh token",
					slog.Int(commonerr.StatusCodeKey, statusCode),
					slog.Any(commonerr.ErrKey, err))
				http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
				return
			}
			ctx = context.WithValue(ctx, TokenCtxKey, freshToken)

			err = metadataUsecase.AddUserToken(ctx, metadata.AddUserTokenArgs{
				UserID: userID.Value,
				Token:  freshToken,
			})
			if err != nil {
				statusCode = http.StatusInternalServerError
				slog.Error("unable to add user token",
					slog.Int(commonerr.StatusCodeKey, statusCode),
					slog.Any(commonerr.ErrKey, err))
				http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
				return
			}

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
