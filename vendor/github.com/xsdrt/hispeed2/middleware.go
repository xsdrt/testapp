package hispeed2

import (
	"net/http"
	"strconv"

	"github.com/justinas/nosurf"
)

func (h *HiSpeed2) SessionLoad(next http.Handler) http.Handler {
	//h.InfoLog.Println("SessionLoad called")
	return h.Session.LoadAndSave(next)
}

// Setup CSRF protection using nosurf...
func (h *HiSpeed2) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(h.config.cookie.secure)

	csrfHandler.ExemptGlob("/api/*") //could use this if do not want to validate csrf token for certain domains

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Domain:   h.config.cookie.domain,
	})

	return csrfHandler
}
