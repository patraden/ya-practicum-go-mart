package jwtauth

import (
	"net/http"
)

const cookieMaxAge = 3600

func StoreTokenInCookie(w *http.ResponseWriter, token string) {
	http.SetCookie(*w, &http.Cookie{
		Name:     JWTCookie,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   cookieMaxAge,
	})
}

func StoreTokenInHeader(w *http.ResponseWriter, token string) {
	(*w).Header().Add(`Authorization`, `Bearer `+token)
}
