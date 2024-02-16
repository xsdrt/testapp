package hispeed2

import "net/http"

func (h *HiSpeed2) SessionLoad(next http.Handler) http.Handler {
	h.InfoLog.Println("SessionLoad called")
	return h.Session.LoadAndSave(next)
}
