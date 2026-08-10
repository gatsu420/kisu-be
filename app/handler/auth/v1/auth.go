package authhandlerv1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gatsu420/kisu-be/app/usecase/user"
	"github.com/gatsu420/kisu-be/common/commonerr"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

func (h *handlerImpl) GetPermission(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	state := uuid.New().String()
	h.stateRepo.Save(state)

	permissionLink := h.googleAuth.GetPermissionLink(state)
	http.Redirect(w, r, permissionLink, http.StatusFound)
}

func (h *handlerImpl) Callback(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	errUrlParam := r.URL.Query().Get("error")
	if errUrlParam != "" {
		slog.Error("auth server denied request",
			slog.Int(commonerr.StatusCodeKey, http.StatusBadRequest))
		return
	}

	state := r.URL.Query().Get("state")
	stateExistence := h.stateRepo.CheckExistence(state)
	if !stateExistence {
		slog.Error("state doesn't exist",
			slog.Int(commonerr.StatusCodeKey, http.StatusBadRequest))
		return
	}

	token, err := h.googleAuth.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		slog.Error("unable to exchange code from auth server",
			slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError))
		return
	}

	email, err := h.getEmail(context.Background(), token)
	if err != nil {
		errMsg = "unable to get email from google auth"
		slog.Error(errMsg, slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
		return
	}

	userResult, err := h.userUsecase.InsertUser(r.Context(), user.InsertUserArgs{
		Email: email,
	})
	if err != nil {
		slog.Error("unable to insert user",
			slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
		return
	}

	err = h.userUsecase.InsertUserToken(r.Context(), user.InsertUserTokenArgs{
		UserID: userResult.UserID,
		Token:  token,
	})
	if err != nil {
		slog.Error("unable to insert user token",
			slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userResult.UserID,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusOK)
}

func (h *handlerImpl) getEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	googleAuthClient := h.googleAuth.Client(ctx, token)
	resp, err := googleAuthClient.Get("https://openidconnect.googleapis.com/v1/userinfo")
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
