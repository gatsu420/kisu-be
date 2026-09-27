package authhandlerv1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"github.com/gatsu420/kisu-be/common/commoncrypto"
	"github.com/gatsu420/kisu-be/common/commonctx"
	"github.com/gatsu420/kisu-be/common/commonerr"
	"github.com/gatsu420/kisu-be/common/commonhttp"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

func (h *handlerImpl) GetPermission(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	state := buildAuthState(buildStateArgs{
		secret: h.hashSecret,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     commonhttp.AuthStateCookieName,
		Value:    state.state,
		Path:     commonhttp.CookiePath,
		MaxAge:   commonhttp.CookieMaxAge,
		HttpOnly: commonhttp.CookieHttpOnly,
		Secure:   commonhttp.CookieSecure,
		SameSite: commonhttp.CookieSameSite,
	})

	permissionLink := h.googleAuth.GetPermissionLink(googleauthadapter.GetPermissionLinkArgs{
		State: state.state,
	})
	http.Redirect(w, r, permissionLink.Link, http.StatusFound)
}

type buildStateArgs struct {
	secret string
}

type buildStateResult struct {
	state string
}

func buildAuthState(args buildStateArgs) buildStateResult {
	id := uuid.New().String()
	prefix := id + "." + strconv.FormatInt(time.Now().UnixMicro(), 10)
	digest := commoncrypto.HashString(commoncrypto.HashStringArgs{
		Secret: args.secret,
		Str:    prefix,
		Salt:   id,
	})
	return buildStateResult{
		state: prefix + "." + digest.Digest,
	}
}

func (h *handlerImpl) Callback(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	errUrlParam := r.URL.Query().Get("error")
	if errUrlParam != "" {
		errMsg = "auth server denied request"
		statusCode = http.StatusUnauthorized
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode))
		http.Error(w, errMsg, statusCode)
		return
	}

	stateCookie, err := r.Cookie(commonhttp.AuthStateCookieName)
	if err != nil {
		errMsg = "unable to get auth_state cookie"
		statusCode = http.StatusUnauthorized
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	stateVerification, err := verifyAuthState(verifyAuthStateArgs{
		secret:   h.hashSecret,
		cookie:   stateCookie.Value,
		urlParam: r.URL.Query().Get("state"),
	})
	if err != nil {
		errMsg = "unable to verify auth state"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	if !stateVerification.isVerified {
		errMsg = "auth state is not verified"
		statusCode = http.StatusUnauthorized
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode))
		http.Error(w, errMsg, statusCode)
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

	email, err := h.getEmail(r.Context(), token.Token)
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

type verifyAuthStateArgs struct {
	secret   string
	cookie   string
	urlParam string
}

type verifyAuthStateResult struct {
	isVerified            bool
	failedVerificationMsg string
}

func verifyAuthState(args verifyAuthStateArgs) (verifyAuthStateResult, error) {
	cookieParts := strings.Split(args.cookie, ".")
	urlParamParts := strings.Split(args.urlParam, ".")
	comparison, err := commoncrypto.VerifyHashedString(commoncrypto.VerifyHashedStringArgs{
		Secret: args.secret,
		Str:    cookieParts[0] + "." + cookieParts[1],
		Salt:   cookieParts[0],
		Digest: urlParamParts[2],
	})
	if err != nil {
		return verifyAuthStateResult{}, err
	}

	if !comparison.IsSameHash {
		return verifyAuthStateResult{
			failedVerificationMsg: "cookie state does not match URL param",
		}, nil
	}

	epoch, err := strconv.Atoi(cookieParts[1])
	if err != nil {
		return verifyAuthStateResult{}, errors.New("unable to cast epoch from cookie state to int")
	}

	if time.Since(time.UnixMicro(int64(epoch))) > 1*time.Minute {
		return verifyAuthStateResult{
			failedVerificationMsg: "auth state is expired",
		}, nil
	}

	return verifyAuthStateResult{
		isVerified: true,
	}, nil
}

func (h *handlerImpl) getEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	googleAuthClient := h.googleAuth.Client(ctx, googleauthadapter.ClientArgs{
		Token: token,
	})

	ctx, cancel := context.WithTimeout(ctx, commonctx.DefaultCtxTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx,
		"GET",
		"https://openidconnect.googleapis.com/v1/userinfo",
		nil)
	if err != nil {
		return "", fmt.Errorf("unable to construct request for user info: %w", err)
	}

	resp, err := googleAuthClient.Client.Do(req)
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
