package authhandlerv1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"github.com/gatsu420/kisu-be/common/commonerr"
	"github.com/gatsu420/kisu-be/common/commonhttp"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

func (h *handlerImpl) GetPermission(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	state := uuid.New().String()
	h.stateRepo.Save(state)

	permissionLink := h.googleAuth.GetPermissionLink(googleauthadapter.GetPermissionLinkArgs{
		State: state,
	})
	http.Redirect(w, r, permissionLink.Link, http.StatusFound)
}

func (h *handlerImpl) Callback(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	errUrlParam := r.URL.Query().Get("error")
	if errUrlParam != "" {
		slog.Error("auth server denied request",
			slog.Int(commonerr.StatusCodeLogKey, http.StatusBadRequest))
		return
	}

	state := r.URL.Query().Get("state")
	stateExistence := h.stateRepo.CheckExistence(state)
	if !stateExistence {
		slog.Error("state doesn't exist",
			slog.Int(commonerr.StatusCodeLogKey, http.StatusBadRequest))
		return
	}

	token, err := h.googleAuth.Exchange(r.Context(), googleauthadapter.ExchangeArgs{
		Code: r.URL.Query().Get("code"),
	})
	if err != nil {
		slog.Error("unable to exchange code from auth server",
			slog.Int(commonerr.StatusCodeLogKey, http.StatusInternalServerError))
		return
	}

	email, err := h.getEmail(context.Background(), token.Token)
	if err != nil {
		errMsg = "unable to get email from google auth"
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeLogKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrLogKey, err))
		return
	}

	addUserResult, err := h.metadataUsecase.AddUser(r.Context(), metadata.AddUserArgs{
		Email: email,
	})
	if err != nil {
		slog.Error("unable to add user",
			slog.Int(commonerr.StatusCodeLogKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrLogKey, err))
		return
	}

	err = h.metadataUsecase.AddUserToken(r.Context(), metadata.AddUserTokenArgs{
		UserID: addUserResult.UserID,
		Token:  token.Token,
	})
	if err != nil {
		slog.Error("unable to add user token",
			slog.Int(commonerr.StatusCodeLogKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrLogKey, err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     commonhttp.UserIDCookieName,
		Value:    addUserResult.UserID,
		Path:     commonhttp.CookiePath,
		MaxAge:   commonhttp.CookieMaxAge,
		HttpOnly: commonhttp.CookieHttpOnly,
		Secure:   commonhttp.CookieSecure,
		SameSite: commonhttp.CookieSameSite,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *handlerImpl) getEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	googleAuthClient := h.googleAuth.Client(ctx, googleauthadapter.ClientArgs{
		Token: token,
	})
	resp, err := googleAuthClient.Client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return "", fmt.Errorf("unable to get user info from google auth: %w", err)
	}
	defer resp.Body.Close()

	var respResult struct {
		Email string `json:"email"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respResult)
	if err != nil {
		return "", fmt.Errorf("unable to decode response body: %w", err)
	}

	return respResult.Email, nil
}
