package config

import (
	"net/http"
	
	"github.com/gorilla/sessions"
)

var AdminStore = sessions.NewCookieStore([]byte("rahasia-admin-99"))
var UserStore = sessions.NewCookieStore([]byte("rahasia-user-88"))

func InitSession() {
	AdminStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	
	UserStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}
