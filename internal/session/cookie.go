package session

import (
	"net/http"
	"time"
)

const (
	SessionDuration = 24 * time.Hour
	SessionIdLength = 32 // 32 bytes = 64 hex chars
)

const (
	CookieName     = "TOKEN"
	cookieMaxAge   = int(24 * time.Hour / time.Second) // 24 hours in seconds
	cookiePath     = "/"
	cookieSecure   = false // TODO: Set to true in production with HTTPS
	cookieHTTPOnly = true
	cookieSameSite = http.SameSiteLaxMode
)

func SetSessionCookie(w http.ResponseWriter, sessionId string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    sessionId,
		Path:     cookiePath,
		MaxAge:   cookieMaxAge,
		Secure:   cookieSecure,
		HttpOnly: cookieHTTPOnly,
		SameSite: cookieSameSite,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     cookiePath,
		MaxAge:   -1, // Immediately expire
		Secure:   cookieSecure,
		HttpOnly: cookieHTTPOnly,
		SameSite: cookieSameSite,
	})
}
