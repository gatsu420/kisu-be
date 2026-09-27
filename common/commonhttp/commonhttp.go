package commonhttp

import "net/http"

const (
	AuthStateCookieName = "auth_state"
	UserIDCookieName    = "user_id"
)

const (
	CookiePath     = "/"
	CookieMaxAge   = 3600
	CookieHttpOnly = true
	CookieSecure   = false
	CookieSameSite = http.SameSiteLaxMode
)
