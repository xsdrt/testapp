package middleware

import (
	"net/http"
)

func (m *middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.App.Session.Exists(r.Context(), "userId") {
			http.Error(w, http.StatusText(401), http.StatusUnauthorized)
		}
	})
}
