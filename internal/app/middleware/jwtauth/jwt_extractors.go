package jwtauth

import (
	"net/http"
	"strings"
)

type TokenExtractor func(r *http.Request) string

func TokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(JWTCookie)
	if err != nil {
		return ``
	}

	return cookie.Value
}

func TokenFromHeader(r *http.Request) string {
	bearer := r.Header.Get(`Authorization`)
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == `BEARER` {
		return bearer[7:]
	}

	return ``
}
