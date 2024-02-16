package hispeed2

import "net/http"

func (h *HiSpeed2) SessionLoad(next http.Handler) http.Handler {
	return h.Session.LoadAndSave(next)
}
