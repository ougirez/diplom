package utils

import (
	"net/http"
	"time"
)

func CreateHttpOnlyCookie(name, value string, ttl int64) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(time.Second * time.Duration(ttl)),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}
