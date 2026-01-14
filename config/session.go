package config

import (
	"net/http"
	
	"github.com/gorilla/sessions"
)

var Store = sessions.NewCookieStore([]byte("kunci-rahasia-sangat-aman"))

func InitSession() {
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}
