package commonhttp

import "net/http"

func NewCookiePlaceholder(name string, value string) *http.Cookie {
	cookie := &http.Cookie{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	cookie.Name = name
	cookie.Value = value

	return cookie
}
