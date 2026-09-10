package commonctx

type ctxKey int

const (
	UserIDCtxKey ctxKey = iota
	FilterCtxKey
	SaltCtxKey
	TokenCtxKey
)
