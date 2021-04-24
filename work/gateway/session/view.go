package session

import (
	"net/http"
	"os"
	"time"

	"github.com/oligoden/chassis/device/view"
)

type View struct {
	view.Default
	secure bool
}

func NewView(w http.ResponseWriter) *View {
	v := &View{}
	v.Default = view.Default{}
	v.Response = w
	if os.Getenv("SECURE") == "true" {
		v.secure = true
	}
	return v
}

func (v View) SetUser(m *Model) {
	if m.Err() != nil {
		v.Error(m)
		return
	}

	v.Response.Header().Set("X_User", m.user)
}

func (v View) SetCookie(m *Model) {
	if m.Err() != nil {
		v.Error(m)
		return
	}

	expire := time.Now().Add(24 * 200 * time.Hour)
	cookie := &http.Cookie{
		Name:     "session",
		Value:    m.session,
		Path:     "/",
		Expires:  expire,
		MaxAge:   0,
		HttpOnly: true,
		Secure:   v.secure,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(v.Response, cookie)
}
