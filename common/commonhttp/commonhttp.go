package commonhttp

import "net/http"

var DefaultCookie *http.Cookie = &http.Cookie{
	Path:     "/",
	MaxAge:   3600,
	HttpOnly: true,
	Secure:   false,
	SameSite: http.SameSiteStrictMode,
}
