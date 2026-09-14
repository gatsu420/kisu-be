package commonctx

import "time"

type ctxKey int

const (
	UserIDCtxKey ctxKey = iota
	FilterCtxKey
	SaltCtxKey
	TokenCtxKey
)

const DefaultCtxTimeout = 5 * time.Second
