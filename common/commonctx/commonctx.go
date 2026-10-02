package commonctx

import "time"

type ctxKey int

const (
	UserIDCtxKey ctxKey = iota
	ParamNameCtxKey
	SaltCtxKey
	TokenCtxKey
)

const DefaultCtxTimeout = 5 * time.Second
