package commonerr

import "errors"

const (
	StatusCodeLogKey = "status_code"
	ErrLogKey        = "err"
)

const UnauthorizedRequestErrMsg = "request is unauthorized"

var NoRowsQueryErr = errors.New("query result has no rows")
