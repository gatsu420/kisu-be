package answerhandlerv1

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gatsu420/kisu-be/app/usecase/answer"
	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"github.com/gatsu420/kisu-be/common/commoncrypto"
	"github.com/gatsu420/kisu-be/common/commonctx"
	"github.com/gatsu420/kisu-be/common/commonerr"
	"github.com/gatsu420/kisu-be/common/commonhttp"
	"github.com/gatsu420/kisu-be/common/commontype"
	"github.com/google/uuid"
)

type RouteToolResult struct {
	Tool json.RawMessage     `json:"tool"`
	Type commontype.ToolType `json:"type"`
}

func (h *handlerImpl) RouteTool(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	userID, ok := r.Context().Value(commonctx.UserIDCtxKey).(*http.Cookie)
	if !ok {
		statusCode = http.StatusUnauthorized
		slog.Error("unable to get user_id cookie",
			slog.Int(commonerr.StatusCodeLogKey, statusCode))
		http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
		return
	}

	salt := uuid.New().String()
	result, err := h.answerUsecase.RouteTool(r.Context(), answer.RouteToolArgs{
		Prompt:     r.URL.Query().Get("prompt"),
		ParamName:  r.URL.Query().Get("param_name"),
		ParamValue: r.URL.Query().Get("param_value"),
		Salt:       salt,
		UserID:     userID.Value,
	})
	if err != nil {
		errMsg = "unable to get answer"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	marshaledResult, err := json.Marshal(result)
	if err != nil {
		errMsg = "unable to marshal result of tool routing"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	hashedResult := buildRoute(buildRouteArgs{
		secret: h.hashSecret,
		str:    string(marshaledResult),
		salt:   salt,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     commonhttp.RouteToolResultCookieName,
		Value:    hashedResult.route,
		Path:     commonhttp.CookiePath,
		MaxAge:   commonhttp.CookieMaxAge,
		HttpOnly: commonhttp.CookieHttpOnly,
		Secure:   commonhttp.CookieSecure,
		SameSite: commonhttp.CookieSameSite,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("tool is routed"))
}

type buildRouteArgs struct {
	secret string
	str    string
	salt   string
}

type buildRouteResult struct {
	route string
}

func buildRoute(args buildRouteArgs) buildRouteResult {
	prefix := base64.URLEncoding.EncodeToString([]byte(
		args.str,
	)) +
		"." + args.salt +
		"." + strconv.FormatInt(time.Now().UnixMicro(), 10)
	digest := commoncrypto.HashString(commoncrypto.HashStringArgs{
		Secret: args.secret,
		Str:    prefix,
		Salt:   args.salt,
	})

	return buildRouteResult{
		route: base64.URLEncoding.EncodeToString([]byte(
			prefix + "." + digest.Digest,
		)),
	}
}

type GetAnswerResult struct {
	Answer               json.RawMessage `json:"answer"`
	StringifiedFuncCalls string          `json:"stringified_func_calls"`
}

func (h *handlerImpl) GetAnswer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	routeCookie, err := r.Cookie(commonhttp.RouteToolResultCookieName)
	if err != nil {
		statusCode = http.StatusUnauthorized
		slog.Error("unable to get route cookie",
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
		return
	}

	decodedCookie, err := base64.URLEncoding.DecodeString(routeCookie.Value)
	if err != nil {
		errMsg = "unable to decode route cookie"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}
	decodedCookieParts := strings.Split(string(decodedCookie), ".")
	routeVerification, err := verifyRoute(verifyRouteArgs{
		secret: h.hashSecret,
		cookie: routeCookie.Value,
	})
	if err != nil {
		errMsg = "unable to verify route"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	if !routeVerification.isVerified {
		if routeVerification.failedVerificationMsg != "" {
			errMsg = routeVerification.failedVerificationMsg
		} else {
			errMsg = commonerr.UnauthorizedRequestErrMsg
		}
		statusCode = http.StatusUnauthorized
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode))
		http.Error(w, errMsg, statusCode)
		return
	}

	decodedRouteStr, err := base64.URLEncoding.DecodeString(decodedCookieParts[0])
	if err != nil {
		errMsg = "unable to decode route str"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}
	var route RouteToolResult
	err = json.Unmarshal([]byte(string(decodedRouteStr)), &route)
	if err != nil {
		errMsg = "unable to unmarshal route cookie parts"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	toolResult, err := h.metadataUsecase.CallTool(r.Context(), metadata.CallToolArgs{
		Salt:   decodedCookieParts[1],
		Tool:   route.Tool,
		Type:   route.Type,
		Limit:  r.URL.Query().Get("limit"),
		Offset: r.URL.Query().Get("offset"),
	})
	if err != nil {
		errMsg = "unable to call tool"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	err = json.NewEncoder(w).Encode(GetAnswerResult{
		Answer:               toolResult.Result,
		StringifiedFuncCalls: toolResult.StringifiedTool,
	})
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
	}
}

type verifyRouteArgs struct {
	secret string
	cookie string
}

type verifyRouteResult struct {
	isVerified            bool
	failedVerificationMsg string
}

func verifyRoute(args verifyRouteArgs) (verifyRouteResult, error) {
	decodedCookie, err := base64.URLEncoding.DecodeString(args.cookie)
	if err != nil {
		return verifyRouteResult{}, fmt.Errorf("unable to decode cookie: %w", err)
	}

	cookieParts := strings.Split(string(decodedCookie), ".")
	comparison, err := commoncrypto.VerifyHashedString(commoncrypto.VerifyHashedStringArgs{
		Secret: args.secret,
		Str: cookieParts[0] +
			"." + cookieParts[1] +
			"." + cookieParts[2],
		Salt:   cookieParts[1],
		Digest: cookieParts[3],
	})
	if err != nil {
		return verifyRouteResult{}, err
	}

	if !comparison.IsSameHash {
		return verifyRouteResult{
			failedVerificationMsg: "cookie state does not match URL param",
		}, nil
	}

	epoch, err := strconv.Atoi(cookieParts[2])
	if err != nil {
		return verifyRouteResult{}, errors.New("unable to cast epoch from cookie state to int")
	}

	if time.Since(time.UnixMicro(int64(epoch))) > 1*time.Minute {
		return verifyRouteResult{}, nil
	}

	return verifyRouteResult{
		isVerified: true,
	}, nil
}

type UploadCsvArgs struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type UploadCsvResult struct {
	Url string `json:"url"`
}

func (h *handlerImpl) UploadCsv(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	var args UploadCsvArgs
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		errMsg = "unable to decode request body"
		statusCode = http.StatusBadRequest
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	result, err := h.answerUsecase.UploadCsv(r.Context(),
		answer.UploadCsvArgs{
			Name:    args.Name,
			Content: strings.NewReader(args.Content),
		})
	if err != nil {
		errMsg = "unable to upload csv"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	err = json.NewEncoder(w).Encode(UploadCsvResult{
		Url: result.Url,
	})
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
	}
}

type AddToolArgs struct {
	ToolDescription string              `json:"tool_description"`
	Project         string              `json:"project"`
	Dataset         string              `json:"dataset"`
	TableName       string              `json:"table_name"`
	Columns         []AddToolColumn     `json:"columns"`
	Type            commontype.ToolType `json:"type"`
	Examples        []AddToolExample    `json:"examples"`
	ParamNames      []string            `json:"param_names"`
}

type AddToolColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type AddToolExample struct {
	Description string `json:"description"`
	Query       string `json:"query"`
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
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	columns := []metadata.AddToolColumn{}
	for _, c := range args.Columns {
		columns = append(columns, metadata.AddToolColumn{
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
		})
	}

	examples := []metadata.AddToolExample{}
	for _, e := range args.Examples {
		examples = append(examples, metadata.AddToolExample{
			Description: e.Description,
			Query:       e.Query,
		})
	}

	userID, ok := r.Context().Value(commonctx.UserIDCtxKey).(*http.Cookie)
	if !ok {
		statusCode = http.StatusUnauthorized
		slog.Error("unable to get user_id cookie",
			slog.Int(commonerr.StatusCodeLogKey, statusCode))
		http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
		return
	}

	err = h.metadataUsecase.AddTool(r.Context(), metadata.AddToolArgs{
		UserID:          userID.Value,
		ToolDescription: args.ToolDescription,
		Project:         args.Project,
		Dataset:         args.Dataset,
		TableName:       args.TableName,
		Columns:         columns,
		Type:            args.Type,
		Examples:        examples,
		ParamNames:      args.ParamNames,
	})
	if err != nil {
		errMsg = "unable to add tool"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("tool is added"))
}

type GetToolResult struct {
	Rows []GetToolRow `json:"rows"`
}

type GetToolRow struct {
	ToolDescription string              `json:"tool_description"`
	Project         string              `json:"project"`
	Dataset         string              `json:"dataset"`
	TableName       string              `json:"table_name"`
	Columns         []GetToolColumn     `json:"columns"`
	Type            commontype.ToolType `json:"type"`
	Examples        []GetToolExample    `json:"examples"`
	ParamNames      []string            `json:"param_names"`
}

type GetToolColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type GetToolExample struct {
	Description string `json:"description"`
	Query       string `json:"query"`
}

func (h *handlerImpl) GetTool(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	userID, ok := r.Context().Value(commonctx.UserIDCtxKey).(*http.Cookie)
	if !ok {
		statusCode = http.StatusUnauthorized
		slog.Error("unable to get user_id cookie",
			slog.Int(commonerr.StatusCodeLogKey, statusCode))
		http.Error(w, commonerr.UnauthorizedRequestErrMsg, statusCode)
		return
	}

	tools, err := h.metadataUsecase.GetTool(r.Context(), metadata.GetToolArgs{
		UserID: userID.Value,
	})
	if err != nil {
		errMsg = "unable to get tool"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	resultRows := []GetToolRow{}
	for _, r := range tools.Rows {
		columns := []GetToolColumn{}
		for _, c := range r.Columns {
			columns = append(columns, GetToolColumn{
				Name:        c.Name,
				Type:        c.Type,
				Description: c.Description,
			})
		}

		examples := []GetToolExample{}
		for _, e := range r.Examples {
			examples = append(examples, GetToolExample{
				Description: e.Description,
				Query:       e.Query,
			})
		}

		resultRows = append(resultRows, GetToolRow{
			ToolDescription: r.ToolDescription,
			Project:         r.Project,
			Dataset:         r.Dataset,
			TableName:       r.TableName,
			Columns:         columns,
			Type:            r.Type,
			Examples:        examples,
			ParamNames:      r.ParamNames,
		})
	}

	err = json.NewEncoder(w).Encode(resultRows)
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
	}
}

type ValidateToolQueryResult struct {
	IsValid bool `json:"is_valid"`
}

func (h *handlerImpl) ValidateToolQuery(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	result, err := h.metadataUsecase.ValidateToolQuery(r.Context(), metadata.ValidateToolQueryArgs{
		Query: r.URL.Query().Get("query"),
	})
	if err != nil {
		errMsg = "unable to validate tool query"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	err = json.NewEncoder(w).Encode(ValidateToolQueryResult{
		IsValid: result.IsValid,
	})
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
	}
}

type GetToolTableMetadataResult struct {
	Description string                       `json:"description"`
	Columns     []GetToolTableMetadataColumn `json:"columns"`
}

type GetToolTableMetadataColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func (h *handlerImpl) GetToolTableMetadata(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var errMsg string
	var statusCode int

	result, err := h.metadataUsecase.GetToolTableMetadata(r.Context(), metadata.GetToolTableMetadataArgs{
		Type:         commontype.ToolType(r.URL.Query().Get("type")),
		Project:      r.URL.Query().Get("project"),
		Dataset:      r.URL.Query().Get("dataset"),
		TableName:    r.URL.Query().Get("table_name"),
		BuilderQuery: r.URL.Query().Get("builder_query"),
	})
	if err != nil {
		errMsg = "unable to get tool table metadata"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
		return
	}

	columns := []GetToolTableMetadataColumn{}
	for _, c := range result.Columns {
		columns = append(columns, GetToolTableMetadataColumn{
			Name:        c.Name,
			Type:        c.Type,
			Description: c.Description,
		})
	}

	err = json.NewEncoder(w).Encode(GetToolTableMetadataResult{
		Description: result.Description,
		Columns:     columns,
	})
	if err != nil {
		errMsg = "unable to write response"
		statusCode = http.StatusInternalServerError
		slog.Error(errMsg,
			slog.Int(commonerr.StatusCodeLogKey, statusCode),
			slog.String(commonerr.ErrLogKey, err.Error()))
		http.Error(w, errMsg, statusCode)
	}
}
