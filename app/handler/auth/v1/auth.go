package authhandlerv1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gatsu420/kisu-be/app/usecase/metadata"
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

	addUserResult, err := h.metadataUsecase.AddUser(r.Context(), metadata.AddUserArgs{
		Email: email,
	})
	if err != nil {
		slog.Error("unable to add user",
			slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
		return
	}

	err = h.metadataUsecase.AddUserToken(r.Context(), metadata.AddUserTokenArgs{
		UserID: addUserResult.UserID,
		Token:  token,
	})
	if err != nil {
		slog.Error("unable to add user token",
			slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    addUserResult.UserID,
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

type AddToolColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	IsSelected  bool   `json:"is_selected"`
}

type AddToolQueryExample struct {
	Description string `json:"description"`
	Query       string `json:"query"`
}

type AddToolArgs struct {
	ToolDescription string                `json:"tool_description"`
	TableName       string                `json:"table_name"`
	Columns         []AddToolColumn       `json:"columns"`
	QueryExamples   []AddToolQueryExample `json:"query_examples"`
}

func (h *handlerImpl) AddTool(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	var args AddToolArgs
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		errMsg = "unable to decode request body"
		statusCode = http.StatusBadRequest
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeKey, statusCode),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, statusCode)
		return
	}

	columns := []metadata.AddToolColumn{}
	for _, c := range args.Columns {
		columns = append(columns, metadata.AddToolColumn{
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
			IsSelected:  c.IsSelected,
		})
	}

	queryExamples := []metadata.AddToolQueryExample{}
	for _, qe := range args.QueryExamples {
		queryExamples = append(queryExamples, metadata.AddToolQueryExample{
			Description: qe.Description,
			Query:       qe.Query,
		})
	}

	err = h.metadataUsecase.AddTool(r.Context(), metadata.AddToolArgs{
		ToolDescription: args.ToolDescription,
		TableName:       args.TableName,
		Columns:         columns,
		QueryExamples:   queryExamples,
	})
	if err != nil {
		errMsg = "unable to add tool"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeKey, statusCode),
			slog.Any(commonerr.ErrKey, err))
		http.Error(w, errMsg, statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("tool is added"))
}
